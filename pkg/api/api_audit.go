package api

import (
	"fmt"
	"net/http"
	"strings"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// auditLogger handles the audit events. This could be overridden with audit
// "providers" that just implement the logr interface.
var auditLogger = logf.Log.WithName("api_audit")

// AuditResult contains information about an audit event from the API router.
type AuditResult struct {
	Allowed     bool
	FromOwner   bool
	Actions     []*v1.APIAction
	Resource    string
	UserSession *v1.JWTClaims
	Request     *http.Request
}

// actions maps allowed values to display strings
var actions = map[bool]string{
	true:  "ALLOWED",
	false: "DENIED",
}

// buildAuditMsg builds a user-friendly audit message to pass to the logger.
func buildAuditMsg(result *AuditResult) string {
	msg := fmt.Sprintf(
		"%s %s",
		actions[result.Allowed],
		result.UserSession.User.GetName(),
	)
	actStrs := make([]string, 0)
	for _, act := range result.Actions {
		if actStr := act.String(); actStr != "" {
			actStrs = append(actStrs, actStr)
		}
	}
	if len(actStrs) > 0 {
		msg = msg + fmt.Sprintf(" => %s", strings.Join(actStrs, ","))
	}
	msg = msg + fmt.Sprintf(" => %s", result.Request.URL.Path)
	if result.FromOwner {
		msg = msg + " (OWNER)"
	}
	return msg
}

// auditLog logs the event with parseable metadata.
func (d *desktopAPI) auditLog(result *AuditResult) {
	if !d.vdiCluster.AuditLogEnabled() {
		return
	}
	msg := buildAuditMsg(result)
	auditLogger.Info(
		msg,
		"Allowed", result.Allowed,
		"User.Name", result.UserSession.User.Name,
		"Request.Path", result.Request.URL.Path,
		"API.Actions", result.Actions,
	)
}
