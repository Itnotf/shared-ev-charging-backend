package controllers

import (
	"net/http"
	"shared-charge/service"
	"shared-charge/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LicensePlateController 车牌号控制器
type LicensePlateController struct {
	licensePlateService *service.LicensePlateService
}

// NewLicensePlateController 创建车牌号控制器实例
func NewLicensePlateController() *LicensePlateController {
	return &LicensePlateController{
		licensePlateService: service.NewLicensePlateService(),
	}
}

// CreateLicensePlateRequest 创建车牌号请求
type CreateLicensePlateRequest struct {
	PlateNumber string `json:"plate_number" binding:"required" example:"京A12345"`
	IsDefault   bool   `json:"is_default" example:"false"`
}

// UpdateLicensePlateRequest 更新车牌号请求
type UpdateLicensePlateRequest struct {
	PlateNumber string `json:"plate_number" binding:"required" example:"京A12345"`
}

// GetUserLicensePlates 获取用户车牌号列表
// @Summary 获取用户车牌号列表
// @Description 获取当前用户的所有车牌号
// @Tags 车牌号管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "success"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/user/license-plates [get]
func (c *LicensePlateController) GetUserLicensePlates(ctx *gin.Context) {
	userModel, ok := utils.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	licensePlates, err := c.licensePlateService.GetUserLicensePlates(userModel.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取车牌号列表失败"})
		return
	}

	// 格式化返回数据
	var result []map[string]interface{}
	for _, plate := range licensePlates {
		result = append(result, plate.FormatLicensePlateInfo())
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}

// CreateLicensePlate 创建车牌号
// @Summary 创建车牌号
// @Description 为当前用户添加新的车牌号
// @Tags 车牌号管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateLicensePlateRequest true "车牌号信息"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/user/license-plates [post]
func (c *LicensePlateController) CreateLicensePlate(ctx *gin.Context) {
	userModel, ok := utils.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	var req CreateLicensePlateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	licensePlate, err := c.licensePlateService.CreateLicensePlate(userModel.ID, req.PlateNumber, req.IsDefault)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "创建车牌号失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": licensePlate.FormatLicensePlateInfo()})
}

// UpdateLicensePlate 更新车牌号
// @Summary 更新车牌号
// @Description 更新指定的车牌号信息
// @Tags 车牌号管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "车牌号ID"
// @Param request body UpdateLicensePlateRequest true "车牌号信息"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "车牌号不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/user/license-plates/{id} [put]
func (c *LicensePlateController) UpdateLicensePlate(ctx *gin.Context) {
	userModel, ok := utils.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	plateIDStr := ctx.Param("id")
	plateID, err := strconv.ParseUint(plateIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "车牌号ID格式错误", "error": err.Error()})
		return
	}

	var req UpdateLicensePlateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	err = c.licensePlateService.UpdateLicensePlate(userModel.ID, uint(plateID), req.PlateNumber)
	if err != nil {
		if err.Error() == "车牌号不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "车牌号不存在", "error": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "更新车牌号失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteLicensePlate 删除车牌号
// @Summary 删除车牌号
// @Description 删除指定的车牌号
// @Tags 车牌号管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "车牌号ID"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "车牌号不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/user/license-plates/{id} [delete]
func (c *LicensePlateController) DeleteLicensePlate(ctx *gin.Context) {
	userModel, ok := utils.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	plateIDStr := ctx.Param("id")
	plateID, err := strconv.ParseUint(plateIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "车牌号ID格式错误", "error": err.Error()})
		return
	}

	err = c.licensePlateService.DeleteLicensePlate(userModel.ID, uint(plateID))
	if err != nil {
		if err.Error() == "车牌号不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "车牌号不存在", "error": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "删除车牌号失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// SetDefaultLicensePlate 设置默认车牌号
// @Summary 设置默认车牌号
// @Description 将指定车牌号设为默认
// @Tags 车牌号管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "车牌号ID"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "车牌号不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/user/license-plates/{id}/set-default [put]
func (c *LicensePlateController) SetDefaultLicensePlate(ctx *gin.Context) {
	userModel, ok := utils.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	plateIDStr := ctx.Param("id")
	plateID, err := strconv.ParseUint(plateIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "车牌号ID格式错误", "error": err.Error()})
		return
	}

	err = c.licensePlateService.SetDefaultLicensePlate(userModel.ID, uint(plateID))
	if err != nil {
		if err.Error() == "车牌号不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "车牌号不存在", "error": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "设置默认车牌号失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "设置成功"})
}
