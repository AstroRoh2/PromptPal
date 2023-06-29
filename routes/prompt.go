package routes

import (
	"net/http"
	"strconv"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

func listProjectPrompts(c *gin.Context) {
	var payload queryPagination
	if err := c.BindQuery(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	idStr, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid project id",
		})
		return
	}

	pid, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	prompts, err := service.
		EntClient.
		Prompt.
		Query().
		Where(prompt.HasProjectWith(project.ID(pid))).
		Where(prompt.IDLT(payload.Cursor)).
		Limit(payload.Limit).
		Order(ent.Desc(prompt.FieldID)).
		All(c)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	count, err := service.
		EntClient.
		Prompt.
		Query().
		Where(prompt.HasProjectWith(project.ID(pid))).
		Count(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ListResponse[*ent.Prompt]{
		Count: count,
		Data:  prompts,
	})
}

func listPrompts(c *gin.Context) {
	// TODO: only admin can list all prompts across projects

	var payload queryPagination
	if err := c.BindQuery(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	prompts, err := service.
		EntClient.
		Prompt.
		Query().
		Where(prompt.IDLT(payload.Cursor)).
		Limit(payload.Limit).
		Order(ent.Desc(prompt.FieldID)).
		All(c)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	count, err := service.EntClient.Prompt.Query().Count(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ListResponse[*ent.Prompt]{
		Count: count,
		Data:  prompts,
	})
}

func getPrompt(c *gin.Context) {
	idStr, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid id",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	prompt, err := service.EntClient.Prompt.Get(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, prompt)
}

type createPromptPayload struct {
	ProjectID   int                     `json:"projectId"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	TokenCount  int                     `json:"tokenCount"`
	Prompts     []schema.PromptRow      `json:"prompts"`
	Variables   []schema.PromptVariable `json:"variables"`
	PublicLevel prompt.PublicLevel      `json:"publicLevel"`
}

func createPrompt(c *gin.Context) {
	var payload createPromptPayload
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	p, err := service.
		EntClient.
		Prompt.
		Create().
		SetName(payload.Name).
		SetDescription(payload.Description).
		SetCreatorID(c.GetInt("uid")).
		SetProjectID(payload.ProjectID).
		SetPrompts(payload.Prompts).
		SetVariables(payload.Variables).
		SetPublicLevel(payload.PublicLevel).
		SetTokenCount(payload.TokenCount).
		Save(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, p)
}

func updatePrompt(c *gin.Context) {
}

func testPrompt(c *gin.Context) {
	var payload createPromptPayload
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	pj, err := service.EntClient.Project.Get(c, payload.ProjectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	res, err := openAIService.Chat(c, pj, payload.Prompts, payload.Variables)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}
