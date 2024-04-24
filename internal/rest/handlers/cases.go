package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
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
	clusterRoutes := router.Group("/cluster")
	clusterRoutes.GET("/:clusterID", h.getClusterAction)
	clusterRoutes.POST("/:clusterID", h.createCaseAction)
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

func (h *Case) getClusterAction(c *gin.Context) {
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

	cluster, err := h.api.CasesService.GetCluster(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &casesv1.GetClusterRequest{
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

	var tasksList []*models.Task
	for _, task := range cluster.Tasks {
		tasksList = append(tasksList, &models.Task{
			ID:          task.Id,
			Title:       task.Title,
			Description: task.Description,
			Status:      models.TaskStatus(task.Status),
			CreatedAt:   task.CreatedAt,
			FormedAt:    task.FormedAt,
			CompletedAt: task.CompletedAt,
			Case: &models.Case{
				ID:       task.Case.Id,
				Title:    task.Case.Title,
				Solution: task.Case.Solution,
			},
			User: &models.User{
				ID:    task.User.Id,
				Email: task.User.Email,
			},
		})
	}

	c.JSON(http.StatusOK, models.GetClusterResponse{
		ID:        cluster.Id,
		Name:      cluster.Name,
		Frequency: cluster.Frequency,
		Cases:     casesList,
		Tasks:     tasksList,
	})
}
