package structs

type SysUser struct {
	ID       uint
	Username string
	Password string
}

func (SysUser) TableName() string {
	return "sys_user"
}
