# go-swagger3 中文使用文档

## 目录

- [为什么使用 go-swagger3](#为什么使用-go-swagger3)
- [解决的痛点](#解决的痛点)
- [和常见 Swagger 工具的区别](#和常见-swagger-工具的区别)
- [安装](#安装)
- [生成文档](#生成文档)
- [基础注解](#基础注解)
- [接口注解](#接口注解)
- [参数](#参数)
- [响应](#响应)
- [Header](#header)
- [Tag 与过滤](#tag-与过滤)
- [结构体 Schema tag](#结构体-schema-tag)
- [枚举](#枚举)
- [安全配置](#安全配置)
- [限制](#限制)


## 前言
- 在项目中，我们经常遇到这样的的问题，给出的的接口中，没有描述。并且在代码中也没有备注。对于接口业务的定义就需要挖掘代码来看这个字段到底干什么用的。（多次提出文档字段描述问题，增加接手，复盘，解读成本）
- 在项目开发阶段，特别是项目前期，接口和相关字段，描述和用途会频繁变更，导致更新了代码，又没有更新yapi，出现断层，导致联调时间拉长，沟通成本增加。
- 作为一个后端开发，写完代码，又要去手工一个一个添加文档，如果手里的活儿很多，还很容易忘记，这是一个二次工作，人肉维护时间成本增加，多人协作的情况下，格式参差不齐，还容易出错。


`go-swagger3`  用 Go 代码注释和结构体 tag 生成 OpenAPI 3.0 文档，输出 JSON 或 YAML 格式的接口规范文件。
本项目的 `go-swagger3` fork了 `https://github.com/parvez3019/go-swagger3` 基于这个项目进行了定制。更符合项目的通用习惯，减少复杂性。

## 为什么使用 go-swagger3

API 文档最常见的问题不是“能不能写出来”，而是“能不能长期保持准确”。在业务快速迭代时，接口参数、响应结构、字段含义、错误码和鉴权方式会频繁变化，如果文档和代码分离维护，很容易出现以下问题：

- 接口已经改了，但 Swagger 文档没有同步，前后端联调才发现不一致。
- 请求参数散落在 handler、DTO、binding tag、validate tag 中，手写文档重复劳动多。
- 统一响应结构很常见，例如 `DataRes{data=xxx}`，但传统注解很难表达每个接口真实的 `data` 类型。
- query 参数经常来自一个请求结构体，逐个写 `@Param` 容易漏字段、漏 required、漏 example。
- 字段说明已经存在于业务 tag 中，例如 `dc` 或 `gorm:"comment:..."`，但文档工具不能复用，导致同一份描述写多遍。
- 多版本、多业务线接口混在同一个项目里，需要按 tag 生成子集文档，而不是每次生成全量文档。

`go-swagger3` 的核心思路是：以 Go 代码为事实来源，通过少量注解补充 OpenAPI 无法从类型系统推断的信息，尽量复用已有结构体 tag，减少文档和代码之间的重复维护。

## 解决的痛点

### 1. 文档跟着代码走

接口的请求体、响应体、枚举、字段校验、示例、描述等信息都从 Go 类型和 tag 中解析。结构体字段变更后，重新生成文档即可同步 schema，减少手写 Swagger 的漂移问题。

### 2. 复用业务结构体 tag

除了常见的 `json`、`form`、`example`、`description`，还支持复用业务项目中已有的 tag：

```go
type User struct {
	ID    int64  `json:"id" dc:"用户 ID" example:"100"`
	Name  string `json:"name" gorm:"comment:用户名称"`
	Email string `json:"email" validate:"required" dc:"邮箱"`
}
```

这样字段名、字段描述、required、example 不需要在注解里重复写一遍。

### 3. 支持请求参数结构体解构

很多 GET 接口会把 query 参数绑定到一个结构体。传统写法需要为每个字段写一行 `@Param`，维护成本高：

```go
// @Param keyword query string false "搜索关键词"
// @Param page query int32 true "页码"
// @Param size query int32 false "每页数量"
```

`go-swagger3` 可以直接解构结构体：

```go
// @Param . query ListUserQuery true "Query params"
```

工具会把 `ListUserQuery` 的字段展开成多个 query 参数，并继承字段上的 `dc`、`example`、`binding:"required"` 等信息。

### 4. 更好表达统一响应结构

业务接口经常使用统一响应体：

```go
type DataRes struct {
	Code int         `json:"code" dc:"业务状态码"`
	Msg  string      `json:"msg" dc:"提示信息"`
	Data interface{} `json:"data" dc:"响应数据"`
}
```

不同接口的 `data` 类型不同。`go-swagger3` 支持在响应注解中覆盖指定字段：

```go
// @Success 200 {object} DataRes{data=User} "ok"
// @Success 200 {object} DataRes{data=[]User} "ok"
// @Success 200 {object} PageRes{data=[]User,total=int64} "ok"
```

这能保留统一响应结构，又能准确表达每个接口真实返回的数据类型。

### 5. 面向工程化生成

`go-swagger3` 支持按 handler 目录扫描、按 tag 过滤、输出 JSON/YAML、schema 名称去包名、OperationId 唯一性校验等能力，更适合放进 CI 或发布流程中自动生成接口文档。

## 和常见 Swagger 工具的区别

| 维度 | go-swagger3 | 常见 Swagger 注解工具 |
| --- | --- | --- |
| 规范版本 | 生成 OpenAPI 3.0。 | 很多工具历史上以 Swagger 2.0 为主，或需要额外配置才能输出 OpenAPI 3。 |
| 使用方式 | Go 类型 + 轻量注解 + 结构体 tag。 | 通常更依赖完整注解，接口参数和 schema 信息容易重复书写。 |
| 请求参数解构 | 支持 `@Param . query SomeStruct`，自动展开结构体字段。 | 通常需要为每个 query 参数分别写注解。 |
| 统一响应体 | 支持 `Wrapper{data=RealType}` 这类字段级覆盖。 | 常见写法只能引用固定响应结构，`interface{}` 或泛型数据字段表达不够准确。 |
| 字段描述来源 | 支持 `description`、`dc`、`gorm:"comment:..."`。 | 通常只识别专用 swagger tag，难以复用已有业务 tag。 |
| required 来源 | 支持 `required`、`binding:"required"`、`validate:"required"`、`json/form` tag 中的 `required`。 | 通常需要额外写 swagger 专用 required 注解。 |
| 文档拆分 | 支持 `--handler-path` 和 `--tag` 过滤生成。 | 有些工具只能按包或全量扫描，业务子集生成不够直接。 |
| 框架绑定 | 通过注解描述路由，不强绑定某个 Web 框架。 | 一些工具与特定框架、路由注册方式或注解风格绑定更深。 |

选择 `go-swagger3` 的场景：

- 项目已经有清晰的 request/response 结构体，希望从代码自动生成 OpenAPI 3 文档。
- 业务大量使用统一响应结构，需要准确描述 `data` 字段的真实类型。
- GET/query 参数通过结构体绑定，希望一次注解完成参数展开。
- 结构体字段已有 `dc`、`gorm comment`、`binding`、`validate` 等 tag，不希望重复维护 swagger 专用描述。
- 需要按业务 tag、版本 tag 或 handler 目录生成部分 API 文档。


## 安装

```bash
go install github.com/aveyuan/go-swagger3@latest
```

如果安装后提示 `command not found: go-swagger3`，确认 `$GOPATH/bin` 已加入 `PATH`：

```bash
export PATH="$HOME/go/bin:$PATH"
```

## 生成文档

在 `go.mod` 所在模块目录执行：

```bash
go-swagger3 --main-file-path main.go
```

常用参数：

| 参数 | 说明 |
| --- | --- |
| `--module-path` | Go module 根目录，工具会在该模块内扫描注释。 |
| `--main-file-path` | main 文件路径。默认会根据模块推断；如果 main 文件不在根目录，需要显式指定。 |
| `--handler-path` | 只扫描指定目录下的 handler 注释。 |
| `--output` | 输出文件路径，默认 `oas.json`。 |
| `--generate-yaml` | 输出 YAML。若 `--output` 以 `.json` 结尾，会自动改成 `.yml`。 |
| `--schema-without-pkg` | schema 名称不带包名；如发生重名，会回退为带包名。 |
| `--tag` | 只生成包含指定 `@Tag` 或 `@Resource` 的接口。 |
| `--debug` | 输出调试日志。 |
| `--strict` | 将 Go 解析 warning 视为 fatal error。 |

main 文件不在模块根目录时：

```bash
go-swagger3 \
  --module-path . \
  --main-file-path ./cmd/server/main.go \
  --output oas.json \
  --schema-without-pkg
```

生成 YAML：

```bash
go-swagger3 --module-path . --output oas.json --generate-yaml
```

只生成某个 tag：

```bash
go-swagger3 --module-path . --tag admin --output admin-api.json
```

Docker 用法：

```bash
docker run -t --rm \
  -v $(pwd):/app \
  -w /app \
  parvez3019/go-swagger3:latest \
  --module-path . \
  --output oas.json
```

## 基础注解

服务级注解通常写在 `main.go` 或 `--main-file-path` 指定文件的注释中。

```go
// @Title User API
// @Version 1.0.0
// @Description User service OpenAPI documentation
// @ContactName API Support
// @ContactEmail support@example.com
// @ContactURL https://example.com/support
// @TermsOfServiceUrl https://example.com/terms
// @LicenseName MIT
// @LicenseURL https://opensource.org/licenses/MIT
// @Server https://api.example.com Production
// @Server http://localhost:8080 Local
// @Security AuthorizationHeader read write
// @SecurityScheme AuthorizationHeader http bearer Input your token
package main
```

必填项：

- `@Title`：接口文档标题。
- `@Version`：接口版本。

如果未配置 `@Server`，会生成默认 server：`/`。

## 接口注解

接口注解写在 handler 函数的 Go doc 注释中。注解大小写不敏感，例如 `@Title`、`@title` 都可以识别。

```go
// @Title Get user detail
// @Description Get user detail by id.
// @OperationId GetUser
// @Tag users
// @Param id path int64 true "User id" "100"
// @Success 200 {object} User "ok"
// @Failure 404 {object} ErrorResponse "not found"
// @Route /users/{id} [get]
func GetUser() {
}
```

支持的接口注解：

| 注解 | 说明 |
| --- | --- |
| `@Title` | OpenAPI operation summary。 |
| `@Description` | OpenAPI operation description。多个 `@Description` 会拼接。 |
| `@OperationId` | operationId，必须全局唯一。 |
| `@Param` | 参数或请求体。 |
| `@Header` | 引用 header 参数结构体。 |
| `@Success` | 成功响应。 |
| `@Failure` | 失败响应。 |
| `@ResponseHeader` | 响应 header。 |
| `@Tag` | 接口 tag。 |
| `@Resource` | 接口 tag，和 `@Tag` 等价。 |
| `@Route` | 路由。 |
| `@Router` | 路由，和 `@Route` 等价。 |

路由格式：

```go
// @Route /users/{id} [get]
// @Router /users [post]
```

支持的 HTTP method：`GET`、`POST`、`PATCH`、`PUT`、`DELETE`、`OPTIONS`、`HEAD`、`TRACE`。

## 参数

### 普通参数

格式：

```text
@Param {name} {in} {goType} {required} "{description}" "{example}"
```

示例：

```go
// @Param id path int64 true "User id" "100"
// @Param keyword query string false "Search keyword" "tom"
// @Param token header string true "Access token"
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `{name}` | 参数名。 |
| `{in}` | 参数位置：`path`、`query`、`form`、`header`、`cookie`、`body`、`file`。 |
| `{goType}` | Go 类型。 |
| `{required}` | `true`、`false`、`required`、`optional`。`path` 参数会强制为 required。 |
| `{description}` | 参数描述，必须使用双引号。 |
| `{example}` | 可选示例值，必须使用双引号。 |

### 请求体

`body` 参数会生成 `requestBody`：

```go
// @Param req body CreateUserRequest true "Create user request"
// @Route /users [post]
func CreateUser() {
}
```

支持结构体、数组、map、`time.Time`、基础类型等：

```go
// @Param req body []CreateUserRequest true "Batch create user request"
// @Param req body map[string]CreateUserRequest true "Users by key"
```

### 表单与文件上传

`form` 和 `file` 参数会生成 `multipart/form-data` 请求体：

```go
// @Param avatar file ignored true "Avatar file"
// @Param nickname form string false "Nickname"
// @Route /users/avatar [post]
func UploadAvatar() {
}
```

`file` 的 Go 类型会被忽略，schema 固定为 `string` + `binary`。

### `.` 变量解构

当 `@Param` 的参数名为 `.` 时，工具会把结构体字段展开成多个参数。适用于 query、header、cookie 等非 body 参数。

```go
type ListUserQuery struct {
	Keyword string `json:"keyword" dc:"搜索关键词" example:"tom"`
	Page    int32  `json:"page" dc:"页码" example:"1" binding:"required"`
	Size    int32  `json:"size" dc:"每页数量" example:"20"`
}

// @Title List users
// @Param . query ListUserQuery true "Query params"
// @Success 200 {object} ListUserResponse "ok"
// @Route /users [get]
func ListUsers() {
}
```

上面的写法会生成三个 query 参数：`keyword`、`page`、`size`。

展开规则：

- 字段名使用结构体字段的 `json` tag；如果存在 `form` tag，会优先使用 `form` tag。
- 字段描述来自 `description`、`dc` 或 `gorm:"comment:..."`。
- 字段示例来自 `example` 或 `override-example`。
- 字段 required 来自 `required`、`binding:"required"`、`validate:"required"` 或 `json/form` tag 中的 `required`。
- `skip:"true"`、`json:"-"`、`form:"-"`、`go-swagger3:"-"` 会跳过字段。

## 响应

### 基础格式

```text
@Success {status} {jsonType} {goType} "{description}"
@Failure {status} {jsonType} {goType} "{description}"
```

示例：

```go
// @Success 200 {object} User "ok"
// @Failure 400 {object} ErrorResponse "bad request"
```

`jsonType` 支持：

- `object` 或 `{object}`
- `array` 或 `{array}`
- `string` 或 `{string}`
- `integer` 或 `{integer}`
- `boolean` 或 `{boolean}`

空响应可以只写状态码和描述：

```go
// @Success 204 "No content"
```

数组响应：

```go
// @Success 200 {array} []User "ok"
```

基础类型响应：

```go
// @Success 200 {string} string "ok"
// @Success 200 {integer} int32 "ok"
```

### `data=xxx` 响应字段覆盖

对于项目中常见的统一响应结构，可以用 `{field=Type}` 覆盖外层结构中的指定字段 schema。

```go
type DataRes struct {
	Code int         `json:"code" dc:"业务状态码"`
	Msg  string      `json:"msg" dc:"提示信息"`
	Data interface{} `json:"data" dc:"响应数据"`
}

type User struct {
	ID   int64  `json:"id" dc:"用户 ID"`
	Name string `json:"name" dc:"用户名称"`
}

// @Success 200 {object} DataRes{data=User} "ok"
// @Route /users/{id} [get]
func GetUser() {
}
```

生成结果会保留 `DataRes` 的外层字段，同时将 `data` 字段替换为 `User` 的 schema。

数组数据：

```go
// @Success 200 {object} DataRes{data=[]User} "ok"
```

带包名和指针也可以解析，空格和 `*` 会被规范化：

```go
// @Success 200 {object} yhttp.DataRes{data= []*model.User} "ok"
```

多个字段覆盖：

```go
// @Success 200 {object} PageRes{data=[]User,total=int64} "ok"
```

注意事项：

- `{data=xxx}` 中的 `data` 必须是外层结构体中的 JSON 字段名。
- 类型可以是当前包类型、带包名类型、数组、map、基础类型。
- 工具会 clone 外层 schema 后再覆盖字段，避免同一个包装类型在多个接口之间互相污染。

## Header

### 请求 Header

请求 header 通常由两部分组成：

- 在 header 结构体上写 `@HeaderParameters`，生成 `components.parameters`。
- 在接口上写 `@Header`，把这些公共参数以 `$ref` 形式挂到当前 operation。

```go
// @HeaderParameters RequestHeaders
type RequestHeaders struct {
	Authorization string `json:"Authorization" skip:"true"`
	Version       string `json:"Client-Version" dc:"客户端版本" example:"1.0.0"`
	Platform      string `json:"Client-Platform" enum:"android,ios,web" example:"web"`
}

// @Header RequestHeaders
// @Route /users [get]
func ListUsers() {
}
```

`@HeaderParameters` 会把结构体字段写入公共 header 参数组件：

```go
// @HeaderParameters RequestHeaders
type RequestHeaders struct {
	Version string `json:"Client-Version" dc:"客户端版本"`
}
```

`@Header RequestHeaders` 不会重复生成参数内容，只会在接口上添加类似下面的引用：

```json
{
  "$ref": "#/components/parameters/Client-Version"
}
```

### 响应 Header

格式：

```text
@ResponseHeader {status} {headerName} {type} "{description}" "{example}"
```

示例：

```go
// @Success 200 {object} LoginResponse "ok"
// @ResponseHeader 200 Set-Cookie string "Access token cookie" "accessToken=xxx; Path=/; HttpOnly"
// @ResponseHeader 200 X-Request-Id string "Request id"
// @Route /login [post]
func Login() {
}
```

说明：

- `{status}` 必须是合法 HTTP 状态码。
- `{type}` 通常使用 `string`、`integer`、`boolean`、`number`。
- `@ResponseHeader` 可以写在对应 `@Success/@Failure` 之前或之后。
- 同一个状态码可以配置多个响应 header。

## Tag 与过滤

`@Tag` 和 `@Resource` 都会写入 OpenAPI operation tags。

```go
// @Tag users
// @Resource admin
// @Route /admin/users [get]
func AdminUsers() {
}
```

如果未指定 tag 且注解值为空，会使用 `others`。

生成时可以用 `--tag` 过滤接口：

```bash
go-swagger3 --module-path . --tag admin --output admin-api.json
```

只有包含 `admin` tag 的接口会进入最终文档。

## 结构体 Schema tag

工具会根据结构体字段类型和 tag 生成 `components.schemas`。

```go
type CreateUserRequest struct {
	Name     string   `json:"name" dc:"用户名称" example:"Tom" binding:"required"`
	Age      int      `json:"age" minimum:"1" maximum:"150" dc:"年龄"`
	Email    string   `json:"email" pattern:"[\\w.]+@[\\w.]"`
	Password string   `json:"password" minLength:"6" maxLength:"64" writeOnly:"true"`
	Roles    []string `json:"roles" uniqueItems:"true" minItems:"1" maxItems:"20"`
	Internal string   `json:"-" skip:"true"`
}
```

### 字段名

字段名优先级：

- 默认使用 Go 字段名。
- 存在 `json:"name"` 时使用 JSON 名称。
- 存在 `form:"name"` 时使用 form 名称，优先级高于 `json`。
- `json:"-"`、`form:"-"`、`go-swagger3:"-"`、`skip:"true"` 会跳过字段。

### 描述

字段描述支持三种来源，优先级如下：

```go
type User struct {
	Name  string `json:"name" description:"用户名称"`
	Phone string `json:"phone" dc:"手机号"`
	Age   int    `json:"age" gorm:"comment:年龄"`
}
```

优先级：

- `description:"..."`。
- `dc:"..."`。
- `gorm:"comment:..."`。

`dc` tag 适合复用业务项目里已经存在的字段说明。

### required

以下写法都会把字段加入 schema 的 `required`：

```go
type Request struct {
	A string `json:"a" required:"true"`
	B string `json:"b,required"`
	C string `json:"c" binding:"required"`
	D string `json:"d" validate:"required"`
}
```

### example

```go
type Request struct {
	Name  string `json:"name" example:"Tom"`
	Age   int    `json:"age" example:"18"`
	Meta  Meta   `json:"meta" override-example:"{\"source\":\"app\"}"`
	Items []Item `json:"items" example:"[{\"id\":1}]"`
}
```

说明：

- `example` 会按字段 schema 类型转换：boolean、integer、number、array、object 会尝试解析为对应 JSON 类型。
- `override-example` 直接覆盖示例值，适合复杂对象或引用类型。
- 引用类型字段设置 example 后，会移除 `$ref`，以便 OpenAPI 能放入 inline example。

### 支持的 tag

| tag | 类型 | 说明 |
| --- | --- | --- |
| `type` | string | 覆盖 schema type，并清理默认 `$ref` 和 `items`。 |
| `format` | string | 设置 schema format。 |
| `required` | bool/存在即生效 | 标记 required。 |
| `description` | string | 字段描述。 |
| `dc` | string | 字段描述，`description` 不存在时使用。 |
| `gorm` | string | 包含 `comment:` 时提取为字段描述。 |
| `example` | string | 示例值。 |
| `override-example` | string | 强制覆盖示例值。 |
| `skip` | bool | `skip:"true"` 跳过字段。 |
| `$ref` | string | 指定引用 schema。数组字段会作用到 `items.$ref`。 |
| `enum` | string | 逗号分隔枚举值。 |
| `title` | string | schema title。 |
| `maximum` | number | 最大值。 |
| `exclusiveMaximum` | bool | 是否不包含最大值。 |
| `minimum` | number | 最小值。 |
| `exclusiveMinimum` | bool | 是否不包含最小值。 |
| `maxLength` | uint | 字符串最大长度。 |
| `minLength` | uint | 字符串最小长度。 |
| `pattern` | string | 正则。 |
| `maxItems` | uint | 数组最大元素数。 |
| `minItems` | uint | 数组最小元素数。 |
| `uniqueItems` | bool | 数组元素是否唯一。 |
| `maxProperties` | uint | object 最大属性数。 |
| `minProperties` | uint | object 最小属性数。 |
| `additionalProperties` | bool | 是否允许额外属性。 |
| `nullable` | bool | 是否可为 null。 |
| `readOnly` | bool | 是否只读。 |
| `writeOnly` | bool | 是否只写。 |
| `binding` | string | 包含 `required` 时标记 required。 |
| `validate` | string | 包含 `required` 时标记 required。 |

## 枚举

可以直接在字段上使用 `enum`：

```go
type User struct {
	Platform string `json:"platform" enum:"android,ios,web" example:"web"`
}
```

也可以定义独立枚举 schema，再通过 `$ref` 引用：

```go
// @Enum OrderByEnum
type OrderByEnum struct {
	OrderByEnum string `enum:"nearest,popular,new,highest-rated" example:"popular"`
}

type ListUserQuery struct {
	OrderBy string `json:"order_by" $ref:"OrderByEnum"`
}
```

数组字段引用枚举：

```go
type Request struct {
	Statuses []string `json:"statuses" $ref:"StatusEnum"`
}
```

## 安全配置

先定义安全方案，再使用 `@Security` 应用到全局文档。

```go
// @SecurityScheme AuthorizationHeader http bearer Input your token
// @Security AuthorizationHeader read write
```

支持的 `@SecurityScheme` 类型：

| 类型 | 格式 | 示例 |
| --- | --- | --- |
| `http` | `@SecurityScheme {name} http {scheme} {description}` | `@SecurityScheme Auth http bearer Input your token` |
| `apiKey` | `@SecurityScheme {name} apiKey {in} {name} {description}` | `@SecurityScheme ApiKey apiKey header X-API-Key API key auth` |
| `openIdConnect` | `@SecurityScheme {name} openIdConnect {url} {description}` | `@SecurityScheme OIDC openIdConnect https://example.com/.well-known/openid-configuration OIDC` |
| `oauth2AuthCode` | `@SecurityScheme {name} oauth2AuthCode {authorizationUrl} {tokenUrl}` | `@SecurityScheme OAuth oauth2AuthCode /oauth/authorize /oauth/token` |
| `oauth2Implicit` | `@SecurityScheme {name} oauth2Implicit {authorizationUrl}` | `@SecurityScheme OAuth oauth2Implicit /oauth/authorize` |
| `oauth2ResourceOwnerCredentials` | `@SecurityScheme {name} oauth2ResourceOwnerCredentials {tokenUrl}` | `@SecurityScheme OAuth oauth2ResourceOwnerCredentials /oauth/token` |
| `oauth2ClientCredentials` | `@SecurityScheme {name} oauth2ClientCredentials {tokenUrl}` | `@SecurityScheme OAuth oauth2ClientCredentials /oauth/token` |

OAuth2 scope：

```go
// @SecurityScope OAuth read_user Read user data
// @SecurityScope OAuth write_user Write user data
```

## 完整示例

```go
package handler

type DataRes struct {
	Code int         `json:"code" dc:"业务状态码"`
	Msg  string      `json:"msg" dc:"提示信息"`
	Data interface{} `json:"data" dc:"响应数据"`
}

type ListUserQuery struct {
	Keyword string `json:"keyword" dc:"搜索关键词" example:"tom"`
	Page    int32  `json:"page" dc:"页码" example:"1" binding:"required"`
	Size    int32  `json:"size" dc:"每页数量" example:"20"`
}

type User struct {
	ID   int64  `json:"id" dc:"用户 ID" example:"100"`
	Name string `json:"name" dc:"用户名称" example:"Tom"`
}

// @Title List users
// @Description List users by query params.
// @OperationId ListUsers
// @Tag users
// @Param . query ListUserQuery true "Query params"
// @Success 200 {object} DataRes{data=[]User} "ok"
// @ResponseHeader 200 X-Request-Id string "Request id"
// @Route /users [get]
func ListUsers() {
}
```

## 限制

- 仅支持 Go module 项目。
- 匿名结构体字段不支持；嵌入结构体字段会尽量展开。
- `@Param` 的描述和示例必须使用双引号。
- 路由路径只支持常见字符：字母、数字、`_`、`.`、`/`、`-`、`{}`。
- `@OperationId` 必须全局唯一。
- 统一响应的 `{data=xxx}` 语法只会覆盖外层结构中指定字段的 schema，不会自动创建业务上不存在的字段语义。