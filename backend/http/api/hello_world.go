package api

import (
	"strconv"

	app "github.com/richer421/q-deploy/app/hello_world"
	"github.com/richer421/q-deploy/app/hello_world/vo"
	"github.com/richer421/q-deploy/http/common"

	"github.com/gin-gonic/gin"
)

type HelloWorldAPI struct {
	appSvc *app.AppService
}

func NewHelloWorldAPI() *HelloWorldAPI {
	return &HelloWorldAPI{appSvc: app.NewAppService()}
}

// List @Summary 列表
// @Tags    hello-world
// @Produce json
// @Param   page      query int true "页码" minimum(1)
// @Param   page_size query int true "每页数量" minimum(1) maximum(100)
// @Success 200 {object} common.Response{data=vo.ListResp}
// @Router  /v1/hello-world [get]
func (h *HelloWorldAPI) List(c *gin.Context) {
	var req vo.ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, err)
		return
	}
	resp, err := h.appSvc.List(c.Request.Context(), &req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, resp)
}

// Get @Summary 详情
// @Tags    hello-world
// @Produce json
// @Param   id path int true "ID"
// @Success 200 {object} common.Response{data=vo.HelloWorldResp}
// @Router  /v1/hello-world/{id} [get]
func (h *HelloWorldAPI) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, err)
		return
	}
	resp, err := h.appSvc.Get(c.Request.Context(), uint(id))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, resp)
}

// Create @Summary 创建
// @Tags    hello-world
// @Accept  json
// @Produce json
// @Param   body body vo.CreateReq true "创建参数"
// @Success 200 {object} common.Response{data=vo.HelloWorldResp}
// @Router  /v1/hello-world [post]
func (h *HelloWorldAPI) Create(c *gin.Context) {
	var req vo.CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	resp, err := h.appSvc.Create(c.Request.Context(), &req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, resp)
}

// Update @Summary 更新
// @Tags    hello-world
// @Accept  json
// @Produce json
// @Param   id   path int          true "ID"
// @Param   body body vo.UpdateReq true "更新参数"
// @Success 200 {object} common.Response
// @Router  /v1/hello-world/{id} [put]
func (h *HelloWorldAPI) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var req vo.UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	if err := h.appSvc.Update(c.Request.Context(), uint(id), &req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, nil)
}

// Delete @Summary 删除
// @Tags    hello-world
// @Produce json
// @Param   id path int true "ID"
// @Success 200 {object} common.Response
// @Router  /v1/hello-world/{id} [delete]
func (h *HelloWorldAPI) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, err)
		return
	}
	if err := h.appSvc.Delete(c.Request.Context(), uint(id)); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, nil)
}
