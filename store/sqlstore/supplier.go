package sqlstore

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"
	"github.com/webitel/engine/localization"
	sqltrace "log"
	"os"
	"time"

	"encoding/json"
	"github.com/go-gorp/gorp"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	wlog "github.com/webitel/wlog"
	"sync/atomic"
)

const (
	DB_PING_ATTEMPTS     = 18
	DB_PING_TIMEOUT_SECS = 10
)

const (
	EXIT_CREATE_TABLE = 100
	EXIT_DB_OPEN      = 101
	EXIT_PING         = 102
	EXIT_NO_DRIVER    = 103
)

type SqlSupplierOldStores struct {
	user                    store.UserStore
	calendar                store.CalendarStore
	skill                   store.SkillStore
	agentTeam               store.AgentTeamStore
	agent                   store.AgentStore
	agentSkill              store.AgentSkillStore
	outboundResource        store.OutboundResourceStore
	outboundResourceGroup   store.OutboundResourceGroupStore
	outboundResourceInGroup store.OutboundResourceInGroupStore
	queue                   store.QueueStore
	queueResource           store.QueueResourceStore
	queueSkill              store.QueueSkillStore
	queueHook               store.QueueHookStore
	bucket                  store.BucketStore
	bucketInQueue           store.BucketInQueueStore
	communicationType       store.CommunicationTypeStore
	list                    store.ListStore
	member                  store.MemberStore
	routingSchema           store.RoutingSchemaStore
	routingOutboundCall     store.RoutingOutboundCallStore
	routingVariable         store.RoutingVariableStore
	call                    store.CallStore
	emailProfile            store.EmailProfileStore
	chat                    store.ChatStore
	chatPlan                store.ChatPlanStore
	region                  store.RegionStore
	pauseCause              store.PauseCauseStore
	notification            store.NotificationStore
	trigger                 store.TriggerStore
	auditForm               store.AuditFormStore
	auditRate               store.AuditRateStore
	presetQuery             store.PresetQueryStore
	systemSetting           store.SystemSettingsStore
	schemeVersion           store.SchemeVersionsStore
}

type SqlSupplier struct {
	rrCounter int64
	srCounter int64
	next      store.LayeredStoreSupplier
	master    *gorp.DbMap
	replicas  []*gorp.DbMap
	//searchReplicas []*gorp.DbMap
	oldStores      SqlSupplierOldStores
	settings       *model.SqlSettings
	lockedToMaster bool
}

func NewSqlSupplier(settings model.SqlSettings) *SqlSupplier {
	supplier := &SqlSupplier{
		rrCounter: 0,
		srCounter: 0,
		settings:  &settings,
	}

	supplier.initConnection()

	supplier.oldStores.user = NewSqlUserStore(supplier)
	supplier.oldStores.calendar = NewSqlCalendarStore(supplier)
	supplier.oldStores.skill = NewSqlSkillStore(supplier)
	supplier.oldStores.agentTeam = NewSqlAgentTeamStore(supplier)
	supplier.oldStores.agent = NewSqlAgentStore(supplier)
	supplier.oldStores.agentSkill = NewSqlAgentSkillStore(supplier)
	supplier.oldStores.outboundResource = NewSqlOutboundResourceStore(supplier)
	supplier.oldStores.outboundResourceGroup = NewSqlOutboundResourceGroupStore(supplier)
	supplier.oldStores.outboundResourceInGroup = NewSqlOutboundResourceInGroupStore(supplier)
	supplier.oldStores.queue = NewSqlQueueStore(supplier)
	supplier.oldStores.queueResource = NewSqlQueueResourceStore(supplier)
	supplier.oldStores.queueSkill = NewSqlQueueSkillStore(supplier)
	supplier.oldStores.queueHook = NewSqlQueueHookStore(supplier)
	supplier.oldStores.bucket = NewSqlBucketStore(supplier)
	supplier.oldStores.bucketInQueue = NewSqlBucketInQueueStore(supplier)
	supplier.oldStores.communicationType = NewSqlCommunicationTypeStore(supplier)
	supplier.oldStores.list = NewSqlListStore(supplier)

	supplier.oldStores.member = NewSqlMemberStore(supplier)

	supplier.oldStores.routingSchema = NewSqlRoutingSchemaStore(supplier)
	supplier.oldStores.routingOutboundCall = NewSqlRoutingOutboundCallStore(supplier)
	supplier.oldStores.routingVariable = NewSqlRoutingVariableStore(supplier)

	supplier.oldStores.call = NewSqlCallStore(supplier)
	supplier.oldStores.emailProfile = NewSqlEmailProfileStore(supplier)
	supplier.oldStores.region = NewSqlRegionStore(supplier)
	supplier.oldStores.pauseCause = NewSqlPauseCauseStore(supplier)
	supplier.oldStores.notification = NewSqlNotificationStore(supplier)
	supplier.oldStores.trigger = NewSqlTriggerStore(supplier)
	supplier.oldStores.auditForm = NewSqlAuditFormStore(supplier)
	supplier.oldStores.auditRate = NewSqlAuditRateStore(supplier)
	supplier.oldStores.presetQuery = NewSqlPresetQueryStore(supplier)
	supplier.oldStores.systemSetting = NewSqlSystemSettingsStore(supplier)
	supplier.oldStores.schemeVersion = NewSqlSchemeVersionsStore(supplier)

	// todo deprecated
	supplier.oldStores.chat = NewSqlChatStore(supplier)
	supplier.oldStores.chatPlan = NewSqlChatPlanStore(supplier)

	err := supplier.GetMaster().CreateTablesIfNotExists()
	if err != nil {
		wlog.Critical(fmt.Sprintf("error creating database tables: %v", err))
		time.Sleep(time.Second)
		os.Exit(EXIT_CREATE_TABLE)
	}

	return supplier
}

func (s *SqlSupplier) SetChainNext(next store.LayeredStoreSupplier) {
	s.next = next
}

func (s *SqlSupplier) Next() store.LayeredStoreSupplier {
	return s.next
}

func (s *SqlSupplier) QueryTimeout() int {
	return *s.settings.QueryTimeout
}

func (ss *SqlSupplier) GetAllConns() []*gorp.DbMap {
	all := make([]*gorp.DbMap, len(ss.replicas)+1)
	copy(all, ss.replicas)
	all[len(ss.replicas)] = ss.master
	return all
}

func setupConnection(con_type string, dataSource string, settings *model.SqlSettings) *gorp.DbMap {
	db, err := dbsql.Open(*settings.DriverName, dataSource)
	if err != nil {
		wlog.Critical(fmt.Sprintf("failed to open SQL connection to err:%v", err.Error()))
		time.Sleep(time.Second)
		os.Exit(EXIT_DB_OPEN)
	}

	for i := 0; i < DB_PING_ATTEMPTS; i++ {
		wlog.Info(fmt.Sprintf("pinging SQL %v database", con_type))
		ctx, cancel := context.WithTimeout(context.Background(), DB_PING_TIMEOUT_SECS*time.Second)
		defer cancel()
		err = db.PingContext(ctx)
		if err == nil {
			break
		} else {
			if i == DB_PING_ATTEMPTS-1 {
				wlog.Critical(fmt.Sprintf("failed to ping DB, server will exit err=%v", err))
				time.Sleep(time.Second)
				os.Exit(EXIT_PING)
			} else {
				wlog.Error(fmt.Sprintf("failed to ping DB retrying in %v seconds err=%v", DB_PING_TIMEOUT_SECS, err))
				time.Sleep(DB_PING_TIMEOUT_SECS * time.Second)
			}
		}
	}

	db.SetMaxIdleConns(*settings.MaxIdleConns)
	db.SetMaxOpenConns(*settings.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(*settings.ConnMaxLifetimeMilliseconds) * time.Millisecond)

	var dbmap *gorp.DbMap

	if *settings.DriverName == model.DATABASE_DRIVER_POSTGRES {
		dbmap = &gorp.DbMap{Db: db, TypeConverter: typeConverter{}, Dialect: PostgresJSONDialect{}}
	} else {
		wlog.Critical("failed to create dialect specific driver")
		time.Sleep(time.Second)
		os.Exit(EXIT_NO_DRIVER)
	}

	if settings.Trace {
		dbmap.TraceOn("[SQL]", sqltrace.New(os.Stdout, "", sqltrace.LstdFlags))
	}

	return dbmap
}

func (s *SqlSupplier) initConnection() {
	s.master = setupConnection("master", *s.settings.DataSource, s.settings)

	if len(s.settings.DataSourceReplicas) > 0 {
		s.replicas = make([]*gorp.DbMap, len(s.settings.DataSourceReplicas))
		for i, replica := range s.settings.DataSourceReplicas {
			s.replicas[i] = setupConnection(fmt.Sprintf("replica-%v", i), replica, s.settings)
		}
	}
}

func (ss *SqlSupplier) GetMaster() *gorp.DbMap {
	return ss.master
}

func (ss *SqlSupplier) GetReplica() *gorp.DbMap {
	if len(ss.settings.DataSourceReplicas) == 0 || ss.lockedToMaster {
		return ss.GetMaster()
	}

	rrNum := atomic.AddInt64(&ss.rrCounter, 1) % int64(len(ss.replicas))
	return ss.replicas[rrNum]
}

func (ss *SqlSupplier) DriverName() string {
	return *ss.settings.DriverName
}

func (ss *SqlSupplier) User() store.UserStore {
	return ss.oldStores.user
}

func (ss *SqlSupplier) Calendar() store.CalendarStore {
	return ss.oldStores.calendar
}

func (ss *SqlSupplier) Skill() store.SkillStore {
	return ss.oldStores.skill
}

func (ss *SqlSupplier) AgentTeam() store.AgentTeamStore {
	return ss.oldStores.agentTeam
}

func (ss *SqlSupplier) Agent() store.AgentStore {
	return ss.oldStores.agent
}

func (ss *SqlSupplier) AgentSkill() store.AgentSkillStore {
	return ss.oldStores.agentSkill
}

func (ss *SqlSupplier) OutboundResource() store.OutboundResourceStore {
	return ss.oldStores.outboundResource
}

func (ss *SqlSupplier) OutboundResourceGroup() store.OutboundResourceGroupStore {
	return ss.oldStores.outboundResourceGroup
}

func (ss *SqlSupplier) OutboundResourceInGroup() store.OutboundResourceInGroupStore {
	return ss.oldStores.outboundResourceInGroup
}

func (ss *SqlSupplier) RoutingSchema() store.RoutingSchemaStore {
	return ss.oldStores.routingSchema
}

func (ss *SqlSupplier) RoutingOutboundCall() store.RoutingOutboundCallStore {
	return ss.oldStores.routingOutboundCall
}

func (ss *SqlSupplier) RoutingVariable() store.RoutingVariableStore {
	return ss.oldStores.routingVariable
}

func (ss *SqlSupplier) Queue() store.QueueStore {
	return ss.oldStores.queue
}

func (ss *SqlSupplier) QueueResource() store.QueueResourceStore {
	return ss.oldStores.queueResource
}

func (ss *SqlSupplier) QueueSkill() store.QueueSkillStore {
	return ss.oldStores.queueSkill
}

func (ss *SqlSupplier) QueueHook() store.QueueHookStore {
	return ss.oldStores.queueHook
}

func (ss *SqlSupplier) Bucket() store.BucketStore {
	return ss.oldStores.bucket
}

func (ss *SqlSupplier) BucketInQueue() store.BucketInQueueStore {
	return ss.oldStores.bucketInQueue
}

func (ss *SqlSupplier) CommunicationType() store.CommunicationTypeStore {
	return ss.oldStores.communicationType
}

func (ss *SqlSupplier) Member() store.MemberStore {
	return ss.oldStores.member
}

func (ss *SqlSupplier) List() store.ListStore {
	return ss.oldStores.list
}

func (ss *SqlSupplier) Call() store.CallStore {
	return ss.oldStores.call
}

func (ss *SqlSupplier) EmailProfile() store.EmailProfileStore {
	return ss.oldStores.emailProfile
}

func (ss *SqlSupplier) Chat() store.ChatStore {
	return ss.oldStores.chat
}

func (ss *SqlSupplier) ChatPlan() store.ChatPlanStore {
	return ss.oldStores.chatPlan
}

func (ss *SqlSupplier) Region() store.RegionStore {
	return ss.oldStores.region
}

func (ss *SqlSupplier) PauseCause() store.PauseCauseStore {
	return ss.oldStores.pauseCause
}

func (ss *SqlSupplier) Notification() store.NotificationStore {
	return ss.oldStores.notification
}

func (ss *SqlSupplier) Trigger() store.TriggerStore {
	return ss.oldStores.trigger
}

func (ss *SqlSupplier) AuditForm() store.AuditFormStore {
	return ss.oldStores.auditForm
}

func (ss *SqlSupplier) AuditRate() store.AuditRateStore {
	return ss.oldStores.auditRate
}

func (ss *SqlSupplier) PresetQuery() store.PresetQueryStore {
	return ss.oldStores.presetQuery
}

func (ss *SqlSupplier) SystemSettings() store.SystemSettingsStore {
	return ss.oldStores.systemSetting
}

func (ss *SqlSupplier) SchemeVersion() store.SchemeVersionsStore {
	return ss.oldStores.schemeVersion
}

type typeConverter struct{}

func (me typeConverter) ToDb(val interface{}) (interface{}, error) {

	switch t := val.(type) {
	case model.StringMap:
		return model.MapToJson(t), nil
	case map[string]string:
		return model.MapToJson(model.StringMap(t)), nil
	case model.StringArray:
		return model.ArrayToJson(t), nil
	case model.StringInterface:
		return model.StringInterfaceToJson(t), nil
	case map[string]interface{}:
		return model.StringInterfaceToJson(model.StringInterface(t)), nil
	}

	return val, nil
}

func (me typeConverter) FromDb(target interface{}) (gorp.CustomScanner, bool) {
	switch target.(type) {
	case *[]model.MemberCommunication,
		*model.MemberCommunication,
		**model.MemberCommunication,
		**model.CCTask,
		*model.Endpoint,
		**model.Endpoint,
		*model.OutboundResourceParameters,
		**model.QueueTaskProcessing,
		*[]model.Lookup,
		*[]model.AggregateData,
		*[]*model.Lookup,
		*[]*model.AgentInQueueStats,
		*[]*model.CallFile,
		*[]*model.CallAnnotation,
		*[]*model.CallHold,
		*[]*model.ChatMember,
		*[]*model.ChatMessage,
		*[]*model.CCTask,
		*[]model.OutboundResourceGroupTime,
		*[]model.CalendarAcceptOfDay,
		*[]model.AgentChannel,
		*[]*model.QueueReportGeneral,
		*model.QueueAgentAgg,
		*model.AgentChannel,
		**model.Variables,
		**map[string]interface{},
		*[]*model.HistoryFileJob,
		*[]*model.CallFileTranscriptLookup,
		*[]*model.CalendarExceptDate,
		*model.AppointmentProfile,
		*[]model.AppointmentDate,
		*model.StringInterface,
		**model.StringMap,
		**model.StringInterface,
		*model.EavesdropInfo,
		*model.Questions,
		*model.QuestionAnswers,
		*model.MailProfileParams:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*[]byte)
			if !ok {
				return errors.New(localization.T("store.sql.convert_member_communication")) // fixme json
			}
			if *s == nil {
				return nil
			}
			return json.Unmarshal(*s, target)
		}
		return gorp.CustomScanner{Holder: &[]byte{}, Target: target, Binder: binder}, true

	case *model.Lookup:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(localization.T("store.sql.convert_lookup"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true

	case **model.Lookup:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*[]byte)
			if !ok {
				return errors.New(localization.T("store.sql.convert_lookup"))
			}
			if *s == nil {
				return nil
			}
			return json.Unmarshal(*s, target)
		}
		return gorp.CustomScanner{Holder: new([]byte), Target: target, Binder: binder}, true
	case *model.StringMap:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(localization.T("store.sql.convert_string_map"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	case *map[string]string:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(localization.T("store.sql.convert_string_map"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true
	case *model.StringArray,
		**model.StringArray:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*[]byte)
			if !ok {
				return errors.New(localization.T("store.sql.convert_string_array"))
			}

			if *s == nil {
				return nil
			}

			var a pq.StringArray

			if err := a.Scan(*s); err != nil {
				return err
			} else {
				*(target).(*model.StringArray) = model.StringArray(a)
				return nil
			}
		}
		return gorp.CustomScanner{Holder: &[]byte{}, Target: target, Binder: binder}, true
	//case *model.StringInterface:
	//	binder := func(holder, target interface{}) error {
	//		s, ok := holder.(*string)
	//		if !ok {
	//			return errors.New(localization.T("store.sql.convert_string_interface"))
	//		}
	//		b := []byte(*s)
	//		return json.Unmarshal(b, target)
	//	}
	//	return gorp.CustomScanner{Holder: model.StringInterface{}, Target: target, Binder: binder}, true
	case *map[string]interface{}:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New(localization.T("store.sql.convert_string_interface"))
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{Holder: new(string), Target: target, Binder: binder}, true

	case *model.Int64Array:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*[]byte)
			if !ok {
				return errors.New(localization.T("store.sql.convert_int64_array"))
			}
			var a pq.Int64Array

			if err := a.Scan(*s); err != nil {
				return err
			} else {
				*(target).(*model.Int64Array) = model.Int64Array(a)
				return nil
			}
		}
		return gorp.CustomScanner{Holder: new([]byte), Target: target, Binder: binder}, true
	}

	return gorp.CustomScanner{}, false
}
