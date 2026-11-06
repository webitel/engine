package fsvar

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ErrBadName  = errors.New("fsvar: invalid variable name")
	ErrDenied   = errors.New("fsvar: variable name is denied")
	ErrBadValue = errors.New("fsvar: invalid variable value")
)

var (
	nameRegx   = regexp.MustCompile(`^[A-Za-z0-9_.\-]{1,64}$`)
	numberRegx = regexp.MustCompile(`^[0-9+*#A-Za-z.@_\-]{1,128}$`)
)

var (
	denyPrefixes = []string{
		"api_", "execute_on_", "exec_", "wbt_",
		"bridge_", "origination_", "originate_",
		"session_in_hangup", "nolocal", "record_post_process_exec", "sip_execute", "rtp_",
	}
	denyExact = map[string]struct{}{
		"export_vars": {}, "process_cdr": {}, "dialplan": {}, "context": {},
		"park_after_bridge": {}, "transfer_after_bridge": {},
		"hangup_after_bridge": {}, "continue_on_fail": {},
		"proxy_media": {}, "bypass_media": {}, "presence_id": {},
		"api_hangup_hook": {}, "session_in_hangup_hook": {},
		"playback_delimiter": {},
	}
)

func ValidateVariables(vars map[string]string) error {
	for k, v := range vars {
		if err := ValidateUserVariableName(k); err != nil {
			return fmt.Errorf("%q %w", k, err)
		}

		if err := ValidateValue(v); err != nil {
			return fmt.Errorf("%q %w", v, err)
		}
	}

	return nil
}

func ValidateUserVariableName(name string) error {
	if !nameRegx.MatchString(name) {
		return ErrBadName
	}

	l := strings.ToLower(name)
	if _, bad := denyExact[l]; bad {
		return ErrDenied
	}

	for _, p := range denyPrefixes {
		if strings.HasPrefix(l, p) {
			return ErrDenied
		}
	}

	if strings.HasPrefix(l, "sip_") && !strings.HasPrefix(l, "sip_h_x-") {
		return ErrDenied
	}

	return nil
}

func ValidateValue(v string) error {
	if !utf8.ValidString(v) {
		return ErrBadValue
	}

	for _, r := range v {
		if r < 0x20 || r == 0x7f {
			return ErrBadValue
		}
	}

	if strings.Contains(v, "${") || strings.Contains(v, `\`) {
		return ErrBadValue
	}

	return nil
}

func Quote(v string) (string, error) {
	if err := ValidateValue(v); err != nil {
		return "", err
	}

	return "'" + strings.ReplaceAll(v, "'", `\'`) + "'", nil
}

func ValidateNumber(number string) error {
	if !numberRegx.MatchString(number) {
		return ErrBadValue
	}

	return nil
}
