package test

// @Title 获取用户列表
// @Param page query int true "分页"
// @Param limit query int true "大小"
// @Param userid query int true "用户ID"
// @Success 200 object ListDataRes{list=ListUserResp,page=Page} "返回结果"
// @Tag v1.3.9
// @Route /api/list_user [get]
func ListUser() {
}

// @Title 获取用户列表2
// @Param . query PageQuery true "分页"
// @Success 200 object DataRes{data=ListUserResp,page=Page} "返回结果"
// @Tag v1.3.9
// @Route /api/list_user2 [get]
func GetUser() {
}

// @Title 获取用户详情
// @Param id path int true "用户ID"
// @Success 200 object UserDetailResp "返回结果"
// @Failure 404 object ApiError "用户不存在"
// @Tag v1.3.9
// @Route /api/users/{id} [get]
func GetUserDetail() {
}

// @Title 创建用户
// @Param body body CreateUserReq true "创建请求"
// @Success 201 object DataRes{data=UserDetailResp} "创建成功"
// @Failure 400 object ApiError "参数错误"
// @Tag v1.3.9
// @Route /api/users [post]
func CreateUser() {
}

// @Title 更新用户
// @Param id path int true "用户ID"
// @Param body body UpdateUserReq true "更新请求"
// @Success 200 object DataRes{data=UserDetailResp} "更新成功"
// @Failure 400 object ApiError "参数错误"
// @Failure 404 object ApiError "用户不存在"
// @Tag v1.3.9
// @Route /api/users/{id} [put]
func UpdateUser() {
}

// @Title 删除用户
// @Param id path int true "用户ID"
// @Success 204 "删除成功"
// @Failure 404 object ApiError "用户不存在"
// @Tag v1.3.9
// @Route /api/users/{id} [delete]
func DeleteUser() {
}

// @Title 用户分页列表
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Param keyword query string false "关键字"
// @Param status query int false "状态"
// @Success 200 object DataRes{data=PagedUserResp} "返回结果"
// @Failure 400 object ApiError "参数错误"
// @Tag v1.4.0
// @Route /api/users/search [get]
func SearchUsers() {
}
