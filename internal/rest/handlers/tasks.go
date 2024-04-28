package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	grpccli "github.com/markgregr/bestHack_support_REST_server/internal/clients/grpc"
	tasksform "github.com/markgregr/bestHack_support_REST_server/internal/rest/forms/tasks"
	"github.com/markgregr/bestHack_support_REST_server/internal/rest/models"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/helper"
	"github.com/markgregr/bestHack_support_REST_server/pkg/rest/response"
	tasksv1 "github.com/markgregr/bestHack_support_protos/gen/go/workflow/tasks"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type Task struct {
	log     *logrus.Logger
	api     *grpccli.Client
	appID   int32
	analURL string
}

func NewTaskHandler(api *grpccli.Client, log *logrus.Logger, appID int32, analURL string) *Task {
	return &Task{
		log:     log,
		api:     api,
		appID:   appID,
		analURL: analURL,
	}
}

func (h *Task) EnrichRoutes(router *gin.Engine) {
	taskRoutes := router.Group("/task")
	taskRoutes.POST("/", h.createTaskAction)
	taskRoutes.POST("/:taskID/status", h.changeTaskStatusAction)
	taskRoutes.GET("/", h.listTasksAction)
	taskRoutes.GET("/:taskID", h.getTaskAction)
	taskRoutes.PUT("/:taskID/case/:caseID", h.AddCaseToTaskAction)
	taskRoutes.PUT("/:taskID", h.AddSolutionToTaskAction)
	taskRoutes.DELETE("/:taskID/case", h.RemoveCaseFromTaskAction)
	taskRoutes.DELETE("/:taskID/solution", h.RemoveSolutionFromTaskAction)
	userRoutes := router.Group("/user")
	userRoutes.GET("/:userID/task", h.listTasksByUserIDAction)
	userRoutes.GET("/", h.listUsersAction)
}

type ClusterRequest struct {
	Message string `json:"message"`
}

type ClusterResponse struct {
	ClusterFrequency int     `json:"cluster_frequency"`
	ClusterIndex     int     `json:"cluster_index"`
	ClusterName      string  `json:"cluster_name"`
	AverageDuration  float64 `json:"average_duration"`
	AverageReaction  float64 `json:"average_reaction"`
}

const delay = 300

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

	requestBody, err := json.Marshal(ClusterRequest{
		Message: form.(*tasksform.CreateTaskForm).Description,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to marshal request body", op)
		response.HandleError(response.NewInternalError(), c)
		return
	}

	resp, err := http.Post(h.analURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.WithError(err).Errorf("%s: failed to send request to anal", op)
		response.HandleError(response.NewInternalError(), c)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to read response body", op)
		response.HandleError(response.NewInternalError(), c)
		return
	}

	var clusterResp ClusterResponse
	err = json.Unmarshal(body, &clusterResp)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to unmarshal response body", op)
		response.HandleError(response.NewInternalError(), c)
		return
	}

	log.Error(clusterResp)
	task, err := h.api.TaskService.CreateTask(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.CreateTaskRequest{
		Title:           form.(*tasksform.CreateTaskForm).Title,
		Description:     form.(*tasksform.CreateTaskForm).Description,
		ClusterIndex:    int64(clusterResp.ClusterIndex),
		ClusterName:     clusterResp.ClusterName,
		Frequency:       int64(clusterResp.ClusterFrequency),
		AverageDuration: float32(clusterResp.AverageDuration),
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to create task", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}
	c.JSON(http.StatusCreated, models.Task{
		ID:          task.Id,
		Title:       task.Title,
		Description: task.Description,
		Solution:    task.Solution,
		Fire:        task.Fire,
		Status:      models.TaskStatus(task.Status),
		CreatedAt:   task.CreatedAt,
		FormedAt:    task.FormedAt,
		CompletedAt: task.CompletedAt,
		Case: &models.Case{
			ID:       task.Case.Id,
			Title:    task.Case.Title,
			Solution: task.Case.Solution,
		},
		Cluster: &models.Cluster{
			ID:        task.Cluster.Id,
			Name:      task.Cluster.Name,
			Frequency: task.Cluster.Frequency,
		},
	})

	// Создаем канал для синхронизации
	timerDone1 := make(chan bool)
	timerDone2 := make(chan bool)

	// Запускаем горутину для ожидания срабатывания таймера
	go func() {
		// Создаем таймер
		timer1 := time.NewTimer(time.Duration(clusterResp.AverageReaction) * time.Second)
		//timer1 := time.NewTimer(time.Duration(60) * time.Second)
		<-timer1.C
		// Таймер истек, отправляем сигнал в канал
		timerDone1 <- true
	}()

	// Запускаем горутину для ожидания срабатывания таймера
	go func() {
		// Создаем таймер
		timer2 := time.NewTimer(time.Duration(clusterResp.AverageReaction+delay) * time.Second)
		//timer2 := time.NewTimer(time.Duration(10) * time.Second)
		<-timer2.C
		// Таймер истек, отправляем сигнал в канал
		timerDone2 <- true
	}()

	// Ожидаем срабатывания таймера или завершения работы контекста
	select {
	case <-timerDone1:
		// Таймер истек, можно вызвать gRPC-сервер
		_, err := h.api.TaskService.FireTask(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.FireTaskRequest{
			TaskId: task.Id,
		})
		if err != nil {
			log.WithError(err).Errorf("%s: failed to change task status", op)
			response.HandleError(response.ResolveError(err), c)
			return
		}
		log.Infof("%s: Task status changed successfully", op)
	case <-ctx.Done():
		// Контекст отменен, необходимо прекратить ожидание
		return
	}
	// Ожидаем срабатывания таймера или завершения работы контекста
	select {
	case <-timerDone2:
		// Таймер истек, можно вызвать gRPC-сервер
		_, err := h.api.TaskService.AppointUserToTask(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.AppointUserToTaskRequest{
			TaskId: task.Id,
		})
		if err != nil {
			log.WithError(err).Errorf("%s: failed to change task status", op)
			response.HandleError(response.ResolveError(err), c)
			return
		}
		log.Infof("%s: Task status changed successfully", op)
	case <-ctx.Done():
		// Контекст отменен, необходимо прекратить ожидание
		return
	}
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

	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse task_id", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	task, err := h.api.TaskService.ChangeTaskStatus(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.ChangeTaskStatusRequest{
		TaskId: taskID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to change task status", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusOK, models.Task{
		ID:          task.Id,
		Title:       task.Title,
		Description: task.Description,
		Solution:    task.Solution,
		Fire:        task.Fire,
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
		Cluster: &models.Cluster{
			ID:        task.Cluster.Id,
			Name:      task.Cluster.Name,
			Frequency: task.Cluster.Frequency,
		},
	})
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
	var tasksList []models.Task

	// Заполнить промежуточную структуру данными из протовской структуры
	for _, task := range tasks.Tasks {
		tasksList = append(tasksList, models.Task{
			ID:          task.Id,
			Title:       task.Title,
			Description: task.Description,
			Fire:        task.Fire,
			Solution:    task.Solution,
			Status:      models.TaskStatus(task.Status),
			CreatedAt:   task.CreatedAt,
			FormedAt:    task.FormedAt,
			CompletedAt: task.CompletedAt,
			Case: &models.Case{
				ID:       task.Case.Id,
				Title:    task.Case.Title,
				Solution: task.Case.Solution,
			},
			Cluster: &models.Cluster{
				ID:        task.Cluster.Id,
				Name:      task.Cluster.Name,
				Frequency: task.Cluster.Frequency,
			},
			User: &models.User{
				ID:    task.User.Id,
				Email: task.User.Email,
			},
		})
	}
	c.JSON(http.StatusOK, tasksList)
}

func (h *Task) getTaskAction(c *gin.Context) {
	const op = "handlers.Task.getTaskAction"
	log := h.log.WithField("operation", op)
	log.Info("get task")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse task_id", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	task, err := h.api.TaskService.GetTask(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.GetTaskRequest{
		TaskId: taskID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to get task", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusOK, models.Task{
		ID:          task.Id,
		Title:       task.Title,
		Description: task.Description,
		Solution:    task.Solution,
		Fire:        task.Fire,
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
		Cluster: &models.Cluster{
			ID:        task.Cluster.Id,
			Name:      task.Cluster.Name,
			Frequency: task.Cluster.Frequency,
		},
	})
}

func (h *Task) AddCaseToTaskAction(c *gin.Context) {
	const op = "handlers.Task.AddCaseToTaskAction"
	log := h.log.WithField("operation", op)
	log.Info("add case to task")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse task_id", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	caseID, err := strconv.ParseInt(c.Param("caseID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse case_id", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	task, err := h.api.TaskService.AddCaseToTask(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.AddCaseToTaskRequest{
		TaskId: taskID,
		CaseId: caseID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to add case to task", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusOK, models.Task{
		ID:          task.Id,
		Title:       task.Title,
		Description: task.Description,
		Solution:    task.Solution,
		Fire:        task.Fire,
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
		Cluster: &models.Cluster{
			ID:        task.Cluster.Id,
			Name:      task.Cluster.Name,
			Frequency: task.Cluster.Frequency,
		},
	})
}

func (h *Task) AddSolutionToTaskAction(c *gin.Context) {
	const op = "handlers.Task.AddCaseToTaskAction"
	log := h.log.WithField("operation", op)
	log.Info("add case to task")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse task_id", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	form, verr := tasksform.NewAddSolutionToTaskForm().ParseAndValidate(c)
	if verr != nil {
		response.HandleError(verr, c)
		return
	}

	task, err := h.api.TaskService.AddSolutionToTask(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.AddSolutionToTaskRequest{
		TaskId:   taskID,
		Solution: form.(*tasksform.AddSolutionToTaskForm).Solution,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to add case to task", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusOK, models.Task{
		ID:          task.Id,
		Title:       task.Title,
		Description: task.Description,
		Solution:    task.Solution,
		Fire:        task.Fire,
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
		Cluster: &models.Cluster{
			ID:        task.Cluster.Id,
			Name:      task.Cluster.Name,
			Frequency: task.Cluster.Frequency,
		},
	})
}

func (h *Task) RemoveCaseFromTaskAction(c *gin.Context) {
	const op = "handlers.Task.RemoveCaseFromTaskAction"
	log := h.log.WithField("operation", op)
	log.Info("remove case from task")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse task_id", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	task, err := h.api.TaskService.RemoveCaseFromTask(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.RemoveCaseFromTaskRequest{
		TaskId: taskID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to remove case from task", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusOK, models.Task{
		ID:          task.Id,
		Title:       task.Title,
		Description: task.Description,
		Solution:    task.Solution,
		Fire:        task.Fire,
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
		Cluster: &models.Cluster{
			ID:        task.Cluster.Id,
			Name:      task.Cluster.Name,
			Frequency: task.Cluster.Frequency,
		},
	})
}

func (h *Task) RemoveSolutionFromTaskAction(c *gin.Context) {
	const op = "handlers.Task.RemoveSolutionFromTaskAction"
	log := h.log.WithField("operation", op)
	log.Info("remove solution from task")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	taskID, err := strconv.ParseInt(c.Param("taskID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse task_id", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	task, err := h.api.TaskService.RemoveSolutionFromTask(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.RemoveSolutionFromTaskRequest{
		TaskId: taskID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to remove solution from task", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	c.JSON(http.StatusOK, models.Task{
		ID:          task.Id,
		Title:       task.Title,
		Description: task.Description,
		Solution:    task.Solution,
		Fire:        task.Fire,
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
		Cluster: &models.Cluster{
			ID:        task.Cluster.Id,
			Name:      task.Cluster.Name,
			Frequency: task.Cluster.Frequency,
		},
	})
}

func (h *Task) listTasksByUserIDAction(c *gin.Context) {
	const op = "handlers.Task.listTasksByUserIDAction"
	log := h.log.WithField("operation", op)
	log.Info("list tasks by user_id")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		log.WithError(err).Errorf("%s: failed to parse user_id", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}

	tasks, err := h.api.TaskService.ListTasksByUserID(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &tasksv1.ListTasksByUserIDRequest{
		UserId: userID,
	})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to list tasks by user_id", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}
	var tasksList []models.Task

	// Заполнить промежуточную структуру данными из протовской структуры
	for _, task := range tasks.Tasks {
		tasksList = append(tasksList, models.Task{
			ID:          task.Id,
			Title:       task.Title,
			Description: task.Description,
			Fire:        task.Fire,
			Solution:    task.Solution,
			Status:      models.TaskStatus(task.Status),
			CreatedAt:   task.CreatedAt,
			FormedAt:    task.FormedAt,
			CompletedAt: task.CompletedAt,
			Case: &models.Case{
				ID:       task.Case.Id,
				Title:    task.Case.Title,
				Solution: task.Case.Solution,
			},
			Cluster: &models.Cluster{
				ID:        task.Cluster.Id,
				Name:      task.Cluster.Name,
				Frequency: task.Cluster.Frequency,
			},
			User: &models.User{
				ID:    task.User.Id,
				Email: task.User.Email,
			},
		})
	}
	c.JSON(http.StatusOK, tasksList)
}

func (h *Task) listUsersAction(c *gin.Context) {
	const op = "handlers.Task.listUsersAction"
	log := h.log.WithField("operation", op)
	log.Info("list users")

	accessToken := helper.ExtractTokenFromHeaders(c)
	if accessToken == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "app_id", fmt.Sprintf("%d", h.appID))

	users, err := h.api.TaskService.ListUsers(metadata.AppendToOutgoingContext(ctx, "access_token", accessToken), &empty.Empty{})
	if err != nil {
		log.WithError(err).Errorf("%s: failed to list users", op)
		response.HandleError(response.ResolveError(err), c)
		return
	}
	var usersList []models.User

	// Заполнить промежуточную структуру данными из протовской структуры
	for _, user := range users.Users {
		usersList = append(usersList, models.User{
			ID:              user.Id,
			Email:           user.Email,
			AvarageDuration: user.AvarageDuration,
		})
	}
	c.JSON(http.StatusOK, usersList)
}
