// 太上老君AI平台本地化API处理器
package localization

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taishanglaojun/core-services/localization"
)

// LocalizationHandler 本地化API处理器
type LocalizationHandler struct {
	service *localization.LocalizationService
}

// NewLocalizationHandler 创建新的本地化API处理器
func NewLocalizationHandler(service *localization.LocalizationService) *LocalizationHandler {
	return &LocalizationHandler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *LocalizationHandler) RegisterRoutes(router *gin.RouterGroup) {
	localization := router.Group("/localization")
	{
		// 用户本地化设置
		localization.GET("/user/:user_id/context", h.GetUserContext)
		localization.PUT("/user/:user_id/context", h.UpdateUserContext)
		localization.POST("/detect", h.DetectUserLocalization)

		// 文本本地化
		localization.POST("/text", h.LocalizeText)
		localization.POST("/text/batch", h.LocalizeTextBatch)

		// 日期时间本地化
		localization.POST("/datetime", h.LocalizeDateTime)
		localization.POST("/datetime/batch", h.LocalizeDateTimeBatch)

		// 数字和货币本地化
		localization.POST("/number", h.LocalizeNumber)
		localization.POST("/currency", h.LocalizeCurrency)
		localization.POST("/currency/convert", h.ConvertCurrency)

		// 地址和联系信息本地化
		localization.POST("/address", h.LocalizeAddress)
		localization.POST("/name", h.LocalizeName)
		localization.POST("/phone", h.LocalizePhoneNumber)

		// 文化信息
		localization.GET("/culture/:culture_code/business-hours", h.GetBusinessHours)
		localization.GET("/culture/:culture_code/holidays", h.GetHolidays)
		localization.POST("/culture/working-day", h.IsWorkingDay)
		localization.POST("/culture/holiday", h.IsHoliday)
		localization.GET("/culture/:culture_code/color/:color", h.GetColorMeaning)
		localization.POST("/culture/taboo-topic", h.IsTabooTopic)
		localization.GET("/culture/:culture_code/food-restrictions", h.GetFoodRestrictions)

		// 支持的本地化选项
		localization.GET("/supported/locales", h.GetSupportedLocales)
		localization.GET("/supported/timezones", h.GetSupportedTimezones)
		localization.GET("/supported/currencies", h.GetSupportedCurrencies)
		localization.GET("/supported/cultures", h.GetSupportedCultures)

		// 管理功能
		localization.POST("/refresh/exchange-rates", h.RefreshExchangeRates)
		localization.POST("/reload/translations", h.ReloadTranslations)
		localization.GET("/stats", h.GetLocalizationStats)

		// 验证功能
		localization.POST("/validate", h.ValidateLocalizationSettings)
	}
}

// GetUserContext 获取用户本地化上下文
func (h *LocalizationHandler) GetUserContext(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userContext, err := h.service.GetUserContext(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userContext})
}

// UpdateUserContext 更新用户本地化上下文
func (h *LocalizationHandler) UpdateUserContext(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var userContext localization.UserLocalizationContext
	if err := c.ShouldBindJSON(&userContext); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userContext.UserID = userID
	if err := h.service.UpdateUserContext(c.Request.Context(), &userContext); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User context updated successfully"})
}

// DetectUserLocalization 检测用户本地化设置
func (h *LocalizationHandler) DetectUserLocalization(c *gin.Context) {
	var req struct {
		AcceptLanguage string `json:"accept_language"`
		UserAgent      string `json:"user_agent"`
		IPAddress      string `json:"ip_address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userContext, err := h.service.DetectUserLocalization(c.Request.Context(), req.AcceptLanguage, req.UserAgent, req.IPAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userContext})
}

// LocalizeText 本地化文本
func (h *LocalizationHandler) LocalizeText(c *gin.Context) {
	var req struct {
		Key         string                 `json:"key" binding:"required"`
		Params      map[string]interface{} `json:"params"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	localizedContent, err := h.service.LocalizeText(c.Request.Context(), req.Key, req.Params, &req.UserContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": localizedContent})
}

// LocalizeTextBatch 批量本地化文本
func (h *LocalizationHandler) LocalizeTextBatch(c *gin.Context) {
	var req struct {
		Items []struct {
			Key    string                 `json:"key" binding:"required"`
			Params map[string]interface{} `json:"params"`
		} `json:"items" binding:"required"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var results []localization.LocalizedContent
	for _, item := range req.Items {
		localizedContent, err := h.service.LocalizeText(c.Request.Context(), item.Key, item.Params, &req.UserContext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		results = append(results, *localizedContent)
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

// LocalizeDateTime 本地化日期时间
func (h *LocalizationHandler) LocalizeDateTime(c *gin.Context) {
	var req struct {
		DateTime    time.Time `json:"datetime" binding:"required"`
		Format      string    `json:"format"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Format == "" {
		req.Format = "2006-01-02 15:04:05"
	}

	localizedDateTime, err := h.service.LocalizeDateTime(c.Request.Context(), req.DateTime, req.Format, &req.UserContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": localizedDateTime})
}

// LocalizeDateTimeBatch 批量本地化日期时间
func (h *LocalizationHandler) LocalizeDateTimeBatch(c *gin.Context) {
	var req struct {
		Items []struct {
			DateTime time.Time `json:"datetime" binding:"required"`
			Format   string    `json:"format"`
		} `json:"items" binding:"required"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var results []localization.LocalizedDateTime
	for _, item := range req.Items {
		format := item.Format
		if format == "" {
			format = "2006-01-02 15:04:05"
		}

		localizedDateTime, err := h.service.LocalizeDateTime(c.Request.Context(), item.DateTime, format, &req.UserContext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		results = append(results, *localizedDateTime)
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

// LocalizeNumber 本地化数字
func (h *LocalizationHandler) LocalizeNumber(c *gin.Context) {
	var req struct {
		Number      float64 `json:"number" binding:"required"`
		Type        string  `json:"type"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type == "" {
		req.Type = "number"
	}

	localizedNumber, err := h.service.LocalizeNumber(c.Request.Context(), req.Number, req.Type, &req.UserContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": localizedNumber})
}

// LocalizeCurrency 本地化货币
func (h *LocalizationHandler) LocalizeCurrency(c *gin.Context) {
	var req struct {
		Amount       float64 `json:"amount" binding:"required"`
		FromCurrency string  `json:"from_currency" binding:"required"`
		ToCurrency   string  `json:"to_currency" binding:"required"`
		UserContext  localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	localizedCurrency, err := h.service.LocalizeCurrency(c.Request.Context(), req.Amount, req.FromCurrency, req.ToCurrency, &req.UserContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": localizedCurrency})
}

// ConvertCurrency 货币转换
func (h *LocalizationHandler) ConvertCurrency(c *gin.Context) {
	var req struct {
		Amount       float64 `json:"amount" binding:"required"`
		FromCurrency string  `json:"from_currency" binding:"required"`
		ToCurrency   string  `json:"to_currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 这里需要调用货币管理器的转换功能
	// 暂时返回模拟数据
	result := map[string]interface{}{
		"original_amount":   req.Amount,
		"from_currency":     req.FromCurrency,
		"to_currency":       req.ToCurrency,
		"converted_amount":  req.Amount * 1.1, // 模拟汇率
		"exchange_rate":     1.1,
		"conversion_time":   time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// LocalizeAddress 本地化地址
func (h *LocalizationHandler) LocalizeAddress(c *gin.Context) {
	var req struct {
		AddressData map[string]string `json:"address_data" binding:"required"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	localizedAddress, err := h.service.LocalizeAddress(c.Request.Context(), req.AddressData, &req.UserContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": localizedAddress})
}

// LocalizeName 本地化姓名
func (h *LocalizationHandler) LocalizeName(c *gin.Context) {
	var req struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		MiddleName  string `json:"middle_name"`
		Honorific   string `json:"honorific"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	localizedName, err := h.service.LocalizeName(c.Request.Context(), req.FirstName, req.LastName, req.MiddleName, req.Honorific, &req.UserContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{"formatted_name": localizedName}})
}

// LocalizePhoneNumber 本地化电话号码
func (h *LocalizationHandler) LocalizePhoneNumber(c *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	localizedPhone, err := h.service.LocalizePhoneNumber(c.Request.Context(), req.PhoneNumber, &req.UserContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{"formatted_phone": localizedPhone}})
}

// GetBusinessHours 获取营业时间
func (h *LocalizationHandler) GetBusinessHours(c *gin.Context) {
	cultureCode := c.Param("culture_code")
	if cultureCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "culture_code is required"})
		return
	}

	userContext := &localization.UserLocalizationContext{Culture: cultureCode}
	businessHours, err := h.service.GetBusinessHours(c.Request.Context(), userContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{"business_hours": businessHours}})
}

// GetHolidays 获取节假日
func (h *LocalizationHandler) GetHolidays(c *gin.Context) {
	cultureCode := c.Param("culture_code")
	if cultureCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "culture_code is required"})
		return
	}

	userContext := &localization.UserLocalizationContext{Culture: cultureCode}
	holidays, err := h.service.GetHolidays(c.Request.Context(), userContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": holidays})
}

// IsWorkingDay 检查是否为工作日
func (h *LocalizationHandler) IsWorkingDay(c *gin.Context) {
	var req struct {
		Date        time.Time `json:"date" binding:"required"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isWorking, err := h.service.IsWorkingDay(c.Request.Context(), req.Date, &req.UserContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]bool{"is_working_day": isWorking}})
}

// IsHoliday 检查是否为节假日
func (h *LocalizationHandler) IsHoliday(c *gin.Context) {
	var req struct {
		Date        time.Time `json:"date" binding:"required"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isHoliday, holiday, err := h.service.IsHoliday(c.Request.Context(), req.Date, &req.UserContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := map[string]interface{}{
		"is_holiday": isHoliday,
	}
	if isHoliday {
		result["holiday"] = holiday
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetColorMeaning 获取颜色含义
func (h *LocalizationHandler) GetColorMeaning(c *gin.Context) {
	cultureCode := c.Param("culture_code")
	color := c.Param("color")

	if cultureCode == "" || color == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "culture_code and color are required"})
		return
	}

	userContext := &localization.UserLocalizationContext{Culture: cultureCode}
	meaning, err := h.service.GetColorMeaning(c.Request.Context(), color, userContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]string{"meaning": meaning}})
}

// IsTabooTopic 检查是否为禁忌话题
func (h *LocalizationHandler) IsTabooTopic(c *gin.Context) {
	var req struct {
		Topic       string `json:"topic" binding:"required"`
		UserContext localization.UserLocalizationContext `json:"user_context" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isTaboo, err := h.service.IsTabooTopic(c.Request.Context(), req.Topic, &req.UserContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": map[string]bool{"is_taboo": isTaboo}})
}

// GetFoodRestrictions 获取饮食限制
func (h *LocalizationHandler) GetFoodRestrictions(c *gin.Context) {
	cultureCode := c.Param("culture_code")
	if cultureCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "culture_code is required"})
		return
	}

	userContext := &localization.UserLocalizationContext{Culture: cultureCode}
	restrictions, err := h.service.GetFoodRestrictions(c.Request.Context(), userContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": restrictions})
}

// GetSupportedLocales 获取支持的语言列表
func (h *LocalizationHandler) GetSupportedLocales(c *gin.Context) {
	locales := h.service.GetSupportedLocales()
	c.JSON(http.StatusOK, gin.H{"data": locales})
}

// GetSupportedTimezones 获取支持的时区列表
func (h *LocalizationHandler) GetSupportedTimezones(c *gin.Context) {
	timezones := h.service.GetSupportedTimezones()
	c.JSON(http.StatusOK, gin.H{"data": timezones})
}

// GetSupportedCurrencies 获取支持的货币列表
func (h *LocalizationHandler) GetSupportedCurrencies(c *gin.Context) {
	currencies := h.service.GetSupportedCurrencies()
	c.JSON(http.StatusOK, gin.H{"data": currencies})
}

// GetSupportedCultures 获取支持的文化列表
func (h *LocalizationHandler) GetSupportedCultures(c *gin.Context) {
	cultures := h.service.GetSupportedCultures()
	c.JSON(http.StatusOK, gin.H{"data": cultures})
}

// RefreshExchangeRates 刷新汇率
func (h *LocalizationHandler) RefreshExchangeRates(c *gin.Context) {
	if err := h.service.RefreshExchangeRates(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exchange rates refreshed successfully"})
}

// ReloadTranslations 重新加载翻译文件
func (h *LocalizationHandler) ReloadTranslations(c *gin.Context) {
	if err := h.service.ReloadTranslations(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Translations reloaded successfully"})
}

// GetLocalizationStats 获取本地化统计信息
func (h *LocalizationHandler) GetLocalizationStats(c *gin.Context) {
	stats := h.service.GetLocalizationStats(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// ValidateLocalizationSettings 验证本地化设置
func (h *LocalizationHandler) ValidateLocalizationSettings(c *gin.Context) {
	var req struct {
		Locale   string `json:"locale" binding:"required"`
		Timezone string `json:"timezone" binding:"required"`
		Currency string `json:"currency" binding:"required"`
		Culture  string `json:"culture" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ValidateLocalizationSettings(req.Locale, req.Timezone, req.Currency, req.Culture); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Localization settings are valid"})
}