package constants

const (
    RoleAdmin  = "admin"
    RoleAuthor = "author"
    RoleReader = "reader"
)

// RoleHierarchy defines the hierarchy of roles and their permissions
var RoleHierarchy = map[string][]string{
    RoleAdmin:  {RoleAdmin, RoleAuthor, RoleReader},
    RoleAuthor: {RoleAuthor, RoleReader},
    RoleReader: {RoleReader},
}
