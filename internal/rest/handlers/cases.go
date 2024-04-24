package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	grpccli "github.com/markgregr/bestHack_support_REST_server/internal/clients/grpc"
	casesform "github.com/markgregr/bestHack_support_REST_server/internal/rest/forms/cases"
	"github.com/markgregr/bestHack_support_REST_server/internal/rest/models"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/helper"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/response"
	casesv1 "github.com/markgregr/bestHack_support_protos/gen/go/workflow/cases"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"net/http"
	"strconv"
)

type Case struct {
	log   *logrus.Entry
	api   *grpccli.Client
	appID int32
}

func NewCaseHandler(api *grpccli.Client, log *logrus.Logger, appID int32) *Case {
	return &Case{
		log:   logrus.NewEntry(log),
		api:   api,
		appID: appID,
	}
}

func (h *Case) EnrichRoutes(router *gin.Engine) {
	casesRoutes := router.Group("/cluster")
	casesRoutes.GET("/", h.listClustersAction)
	casesRoutes.GET("/:clusterID", h.listCasesFromClusterAction)
	casesRoutes.POST("/:clusterID", h.createCaseAction)
	casesRoutes.PUT("/:clusterID/:caseID", h.updateCaseAction)
	casesRoutes.DELETE("/:caseID", h.deleteCaseAction)
}

func (h *Case) createCaseAction(c *gin.Context) {
	const op = "handlers.Case.createCaseAction"
	log := h.log.WithField("operation", op)
	log.Info("create case")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	clusterID, err := strconv.ParseInt(c.Param("clusterID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse clusterID", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	form, verr := casesform.NewCreateCaseForm().ParseAndValidate(c)
	if verr != nil {
		response.HandleError(verr, c)
		return
	}

	caseItem, err := h.api.CasesService.CreateCase(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &casesv1.CreateCaseRequest{
		Title:     form.(*casesform.CreateCaseForm).Title,
		Solution:  form.(*casesform.CreateCaseForm).Solution,
		ClusterId: clusterID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to create case", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusOK, models.Case{
		ID:       caseItem.Id,
		Title:    caseItem.Title,
		Solution: caseItem.Solution,
		Cluster: &models.Cluster{
			ID:        caseItem.Cluster.Id,
			Name:      caseItem.Cluster.Name,
			Frequency: caseItem.Cluster.Frequency,
		},
	})
}

func (h *Case) listCasesFromClusterAction(c *gin.Context) {
	const op = "handlers.Case.getClusterAction"
	log := h.log.WithField("operation", op)
	log.Info("get cluster")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	clusterID, err := strconv.ParseInt(c.Param("clusterID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse clusterID", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	cluster, err := h.api.CasesService.GetCasesFromCluster(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &casesv1.GetCasesFromClusterRequest{
		Id: clusterID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to get cluster", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	var casesList []*models.Case
	for _, caseItem := range cluster.Cases {
		casesList = append(casesList, &models.Case{
			ID:       caseItem.Id,
			Title:    caseItem.Title,
			Solution: caseItem.Solution,
			Cluster: &models.Cluster{
				ID:        caseItem.Cluster.Id,
				Name:      caseItem.Cluster.Name,
				Frequency: caseItem.Cluster.Frequency,
			},
		})
	}

	c.JSON(http.StatusOK, casesList)
}

func (h *Case) listClustersAction(c *gin.Context) {
	const op = "handlers.Case.listClustersAction"
	log := h.log.WithField("operation", op)
	log.Info("list clusters")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	clusters, err := h.api.CasesService.ListClusters(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &empty.Empty{})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to list clusters", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	var clustersList []*models.Cluster
	for _, cluster := range clusters.Clusters {
		clustersList = append(clustersList, &models.Cluster{
			ID:        cluster.Id,
			Name:      cluster.Name,
			Frequency: cluster.Frequency,
		})
	}

	c.JSON(http.StatusOK, clustersList)
}

func (h *Case) updateCaseAction(c *gin.Context) {
	const op = "handlers.Case.updateCaseAction"
	log := h.log.WithField("operation", op)
	log.Info("update case")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	caseID, err := strconv.ParseInt(c.Param("caseID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse caseID", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	form, verr := casesform.NewCreateCaseForm().ParseAndValidate(c)
	if verr != nil {
		response.HandleError(verr, c)
		return
	}

	caseItem, err := h.api.CasesService.UpdateCase(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &casesv1.UpdateCaseRequest{
		Id:       caseID,
		Title:    form.(*casesform.CreateCaseForm).Title,
		Solution: form.(*casesform.CreateCaseForm).Solution,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to update case", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusOK, models.Case{
		ID:       caseItem.Id,
		Title:    caseItem.Title,
		Solution: caseItem.Solution,
		Cluster: &models.Cluster{
			ID:        caseItem.Cluster.Id,
			Name:      caseItem.Cluster.Name,
			Frequency: caseItem.Cluster.Frequency,
		},
	})
}

func (h *Case) deleteCaseAction(c *gin.Context) {
	const op = "handlers.Case.deleteCaseAction"
	log := h.log.WithField("operation", op)
	log.Info("delete case")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	caseID, err := strconv.ParseInt(c.Param("caseID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse caseID", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	_, err = h.api.CasesService.DeleteCase(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &casesv1.DeleteCaseRequest{
		Id: caseID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to delete case", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.Status(http.StatusOK)
}
