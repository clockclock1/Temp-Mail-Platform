package models

const (
	PermMailboxCreate = "mailbox:create"
	PermMailboxRead   = "mailbox:read"
	PermMailboxDelete = "mailbox:delete"
	PermMessageRead   = "message:read"
	PermMessageDelete = "message:delete"
	PermDomainManage  = "domain:manage"
	PermUserManage    = "user:manage"
	PermRoleManage    = "role:manage"
	PermStatsRead     = "stats:read"
	PermConfigManage  = "config:manage"
)

var DefaultPermissionCatalog = []Permission{
	{Key: PermMailboxCreate, Description: "Create temporary mailboxes"},
	{Key: PermMailboxRead, Description: "Read own or managed mailboxes"},
	{Key: PermMailboxDelete, Description: "Delete temporary mailboxes"},
	{Key: PermMessageRead, Description: "Read received messages"},
	{Key: PermMessageDelete, Description: "Delete messages"},
	{Key: PermDomainManage, Description: "Manage receiving domains"},
	{Key: PermUserManage, Description: "Manage users"},
	{Key: PermRoleManage, Description: "Manage roles and permissions"},
	{Key: PermStatsRead, Description: "Read system stats"},
	{Key: PermConfigManage, Description: "Manage runtime config"},
}
