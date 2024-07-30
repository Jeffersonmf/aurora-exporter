package api

import (
	"context"
	"net/http"

	"brandlovrs.exporter/api/controllers"
	"brandlovrs.exporter/pkg/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func StartRouters(ctx context.Context) {
	// Create a new Gin router
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// use ginSwagger middleware to serve the API docs
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "alive",
		})
	})

	// Initialize Controllers
	brandDnbController := controllers.NewBrandDnbController()
	colabCouponController := controllers.NewColabCouponController()
	brandDnbXColabUserController := controllers.NewBrandDnbXColabUserController()
	brandDnbXInfluencerInviteController := controllers.NewBrandDnbXInfluencerInviteController()
	calendarController := controllers.NewCalendarController()
	colabLevelsXColabMissionXColabRewardController := controllers.NewColabLevelsXColabMissionsXColabRewardController()
	colabMissionController := controllers.NewColabMissionController()
	colabMissionsXColabUserController := controllers.NewColabMissionsXColabUserController()
	colabUserController := controllers.NewColabUserController()
	colabUserDataController := controllers.NewColabUserDataController()
	factOrdersController := controllers.NewFactOrdersController()
	missionXCouponInfoController := controllers.NewMissionXCouponInfoController()
	totalCouponSalesController := controllers.NewTotalCouponSalesController()
	vmBrandTotalBalanceController := controllers.NewVBrandTotalBalanceController()
	vmBrandWalletHistoryController := controllers.NewVMBrandWalletHistoryController()
	vmBrandsController := controllers.NewVMBrandWalletHistoryController()
	vmInstagramProfilesBrandsController := controllers.NewVMInstagramProfilesBrandsController()
	vmInstagramProfilesUsersController := controllers.NewVMInstagramProfilesUsersController()
	vmPostMetricsController := controllers.NewVMPostMetricsController()
	vmSNPostHashtagsController := controllers.NewVMsnPostHashtagsController()
	WorkspaceController := controllers.NewWorkspaceController()

	router.POST("/brand_dnb/", brandDnbController.ImportAllRows)
	router.POST("/colab_coupon/", colabCouponController.ImportAllRows)
	router.POST("/brand_dnb_x_colab_user/", brandDnbXColabUserController.ImportAllRows)
	router.POST("/brand_dnb_x_influencer_invite/", brandDnbXInfluencerInviteController.ImportAllRows)
	router.POST("/calendar/", calendarController.ImportAllRows)
	router.POST("/colab_levels_x_colab_mission_x_colab_reward/", colabLevelsXColabMissionXColabRewardController.ImportAllRows)
	router.POST("/colab_mission/", colabMissionController.ImportAllRows)
	router.POST("/colab_mission_x_colab_user/", colabMissionsXColabUserController.ImportAllRows)
	router.POST("/colab_user/", colabUserController.ImportAllRows)
	router.POST("/colab_user_data/", colabUserDataController.ImportAllRows)
	router.POST("/fact_orders/", factOrdersController.ImportAllRows)
	router.POST("/mission_x_coupon_info/", missionXCouponInfoController.ImportAllRows)
	router.POST("/total_coupon_sales/", totalCouponSalesController.ImportAllRows)
	router.POST("/vm_brand_total_balance/", vmBrandTotalBalanceController.ImportAllRows)
	router.POST("/vm_brand_wallet_history/", vmBrandWalletHistoryController.ImportAllRows)
	router.POST("/vm_instagram_profiles_brands/", vmInstagramProfilesBrandsController.ImportAllRows)
	router.POST("/vm_brands/", vmBrandsController.ImportAllRows)
	router.POST("/vm_instagram_profile_users/", vmInstagramProfilesUsersController.ImportAllRows)
	router.POST("/vm_post_metrics/", vmPostMetricsController.ImportAllRows)
	router.POST("/vm_sn_post_hashtags/", vmSNPostHashtagsController.ImportAllRows)
	router.POST("/workspace/", WorkspaceController.ImportAllRows)

	err := router.Run(":8088")
	if err != nil {
		util.Sugar.Errorf("Fail when starting the WebServer.. -> ", zap.String("msg", err.Error()))
	}
}
