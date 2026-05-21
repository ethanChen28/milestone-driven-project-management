package service

import "goal-manager/backend/internal/domain"

type Permission string

type AuthContext struct {
	Role domain.WorkspaceRole
	User string
}

const (
	PermManageIntegration  Permission = "manage_integration"
	PermManageRoadmap      Permission = "manage_roadmap"
	PermManageProject      Permission = "manage_project"
	PermManageMilestone    Permission = "manage_milestone"
	PermManageWorkItem     Permission = "manage_work_item"
	PermManageWorkstream   Permission = "manage_workstream"
	PermSubmitUpdate       Permission = "submit_update"
	PermManageSyncRule     Permission = "manage_sync_rule"
	PermViewDashboard      Permission = "view_dashboard"
	PermManageAlert        Permission = "manage_alert"
	PermManageNotification Permission = "manage_notification"
	PermRunSync            Permission = "run_sync"
)

var rolePermissions = map[domain.WorkspaceRole]map[Permission]bool{
	domain.RoleAdmin: {
		PermManageIntegration:  true,
		PermManageRoadmap:      true,
		PermManageProject:      true,
		PermManageMilestone:    true,
		PermManageWorkItem:     true,
		PermManageWorkstream:   true,
		PermSubmitUpdate:       true,
		PermManageSyncRule:     true,
		PermViewDashboard:      true,
		PermManageAlert:        true,
		PermManageNotification: true,
		PermRunSync:            true,
	},
	domain.RolePortfolioManager: {
		PermManageRoadmap:    true,
		PermManageProject:    true,
		PermManageMilestone:  true,
		PermManageWorkItem:   true,
		PermManageWorkstream: true,
		PermSubmitUpdate:     true,
		PermManageSyncRule:   true,
		PermViewDashboard:    true,
		PermManageAlert:      true,
		PermRunSync:          true,
	},
	domain.RoleProjectOwner: {
		PermManageProject:    true,
		PermManageMilestone:  true,
		PermManageWorkItem:   true,
		PermManageWorkstream: true,
		PermSubmitUpdate:     true,
		PermManageSyncRule:   true,
		PermViewDashboard:    true,
		PermManageAlert:      true,
	},
	domain.RoleContributor: {
		PermManageWorkItem: true,
		PermSubmitUpdate:   true,
		PermViewDashboard:  true,
	},
	domain.RoleViewer: {
		PermViewDashboard: true,
	},
}

func HasPermission(role domain.WorkspaceRole, perm Permission) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	return perms[perm]
}

func CanWrite(role domain.WorkspaceRole) bool {
	return role != domain.RoleViewer
}

func RoleAuth(role domain.WorkspaceRole) AuthContext {
	return AuthContext{Role: role}
}

func ActorAuth(role domain.WorkspaceRole, user string) AuthContext {
	return AuthContext{Role: role, User: user}
}

func (auth AuthContext) HasUser() bool {
	return auth.User != ""
}

func (auth AuthContext) IsAdmin() bool {
	return auth.Role == domain.RoleAdmin
}
