package constants

// User roles
const (
    RoleAdmin  = "admin"
    RoleAuthor = "author"
    RoleReader = "reader"
)

// ValidRoles is a list of all valid user roles
var ValidRoles = []string{RoleAdmin, RoleAuthor, RoleReader}

// RoleHierarchy defines which roles have access to other roles' permissions
var RoleHierarchy = map[string][]string{
    RoleAdmin:  {RoleAdmin, RoleAuthor, RoleReader},
    RoleAuthor: {RoleAuthor, RoleReader},
    RoleReader: {RoleReader},
}
