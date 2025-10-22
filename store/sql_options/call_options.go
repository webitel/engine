package sqloptions

// Custom type representing history call entity additional where filter option.
// Returning function that receive pointer to where filter query string and
// map of query args.
type HistoryCallSQLFilterOption func(*string, map[string]any)

// User history call search where filter option decorating default where query with 
// user with specific grant or calls agent.
// userId - search user id.
// grant - search user grant.
func WithUserGrantFilterOption(userId uint, grant string) HistoryCallSQLFilterOption {
	return func(filterString *string, filterArgs map[string]any) {
		userGlobalSelectFilter := `
			and (
				call_center.cc_user_has_grant(:Domain, :AgentUserId, :Grant)
				or (t.user_id = :AgentUserId or :AgentUserId = any(t.user_ids))
			)
		`
		filterArgs["AgentUserId"] = userId
		filterArgs["Grant"] = grant

		*filterString += userGlobalSelectFilter
	}
}
