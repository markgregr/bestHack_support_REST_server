package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	grpccli "github.com/markgregr/bestHack_support_REST_server/internal/clients/grpc"
	tasksform "github.com/markgregr/bestHack_support_REST_server/internal/rest/forms/tasks"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/helper"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/response"
	tasksv1 "github.com/markgregr/bestHack_support_protos/gen/go/workflow/tasks"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"net/http"
	"strconv"
)

type Task struct {
	log   *logrus.Logger
	api   *grpccli.Client
	appID int32
}

func NewTaskHandler(api *grpccli.Client, log *logrus.Logger, appID int32) *Task {
	return &Task{
		log:   log,
		api:   api,
		appID: appID,
	}
}

func (h *Task) EnrichRoutes(router *gin.Engine) {
	taskRoutes := router.Group("/task")
	taskRoutes.POST("/create", h.createTaskAction)
	taskRoutes.POST("/change-status", h.changeTaskStatusAction)
	taskRoutes.GET("/list", h.listTasksAction)
}

func (h *Task) createTaskAction(c *gin.Context) {
	const op = "handlers.Task.createTaskAction"
	log := h.log.WithField("operation", op)
	log.Info("create task")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	form, verr := tasksform.NewCreateTaskForm().ParseAndValidate(c)
	if verr != nil {
		response.HandleError(verr, c)
		return
	}

	task, err := h.api.TaskService.CreateTask(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.CreateTaskRequest{
		Title:       form.(*tasksform.CreateTaskForm).Title,
		Description: form.(*tasksform.CreateTaskForm).Description,
		ClusterId:   form.(*tasksform.CreateTaskForm).ClusterID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to create task", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *Task) changeTaskStatusAction(c *gin.Context) {
	const op = "handlers.Task.changeTaskStatusAction"
	log := h.log.WithField("operation", op)
	log.Info("change task status")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	form, verr := tasksform.NewChangeTaskStatusForm().ParseAndValidate(c)
	if verr != nil {
		response.HandleError(verr, c)
		return
	}

	task, err := h.api.TaskService.ChangeTaskStatus(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.ChangeTaskStatusRequest{
		TaskId: form.(*tasksform.ChangeTaskStatusForm).TaskID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to change task status", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Task) listTasksAction(c *gin.Context) {
	const op = "handlers.Task.listTasksAction"
	log := h.log.WithField("operation", op)
	log.Info("list tasks")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	status, err := strconv.Atoi(c.Query("status"))
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse status", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	tasks, err := h.api.TaskService.ListTasks(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.ListTasksRequest{
		Status: int64(status),
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to list tasks", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusOK, tasks)
}
