package test

type ListUserResp struct {
	Id   int    `json:"id" dc:"用户ID"`
	Name string `json:"name" dc:"用户名"`
}

type UserDetailResp struct {
	Id        int      `json:"id" dc:"用户ID"`
	Name      string   `json:"name" dc:"用户名"`
	Email     string   `json:"email" dc:"邮箱"`
	Phone     string   `json:"phone" dc:"手机号"`
	Role      string   `json:"role" dc:"角色"`
	Status    int      `json:"status" dc:"状态"`
	Tags      []string `json:"tags" dc:"标签"`
	CreatedAt string   `json:"created_at" dc:"创建时间"`
	UpdatedAt string   `json:"updated_at" dc:"更新时间"`
}

type CreateUserReq struct {
	Name     string `json:"name" dc:"用户名"`
	Email    string `json:"email" dc:"邮箱"`
	Phone    string `json:"phone" dc:"手机号"`
	Password string `json:"password" dc:"密码"`
	Role     string `json:"role" dc:"角色"`
}

type UpdateUserReq struct {
	Id       int    `json:"id" dc:"用户ID"`
	Name     string `json:"name" dc:"用户名"`
	Email    string `json:"email" dc:"邮箱"`
	Phone    string `json:"phone" dc:"手机号"`
	Password string `json:"password" dc:"密码"`
	Status   int    `json:"status" dc:"状态"`
}

type DeleteUserReq struct {
	Id int `json:"id" dc:"用户ID"`
}

type PageQuery struct {
	Page    int    `json:"page" dc:"页码"`
	Limit   int    `json:"limit" dc:"每页数量"`
	Keyword string `json:"keyword" dc:"关键字"`
	Role    string `json:"role" dc:"角色"`
	Status  int    `json:"status" dc:"状态"`
	SortBy  string `json:"sort_by" dc:"排序字段"`
	Order   string `json:"order" dc:"排序方向"`
}

type DataRes struct {
	Code int         `json:"code" dc:"业务状态码"`
	Msg  string      `json:"msg" dc:"提示信息"`
	Data interface{} `json:"data" dc:"响应数据"`
}

type DataResPage struct {
	Code int         `json:"code" dc:"业务状态码"`
	Msg  string      `json:"msg" dc:"提示信息"`
	Data ListDataRes `json:"data" dc:"响应数据"`
}

type ListDataRes struct {
	List interface{} `json:"list" dc:"列表数据"`
	Page interface{} `json:"page" dc:"分页信息"`
}

type Page struct {
	Current int `json:"current" dc:"当前页"`
	Size    int `json:"size" dc:"每页大小"`
	Total   int `json:"total" dc:"总记录数"`
}

type PagedUserResp struct {
	List []ListUserResp `json:"list" dc:"用户列表"`
	Page Page           `json:"page" dc:"分页信息"`
}

type ApiError struct {
	Code    int    `json:"code" dc:"错误码"`
	Message string `json:"message" dc:"错误信息"`
	TraceId string `json:"trace_id" dc:"链路ID"`
}
