package admin

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	adminpb "github.com/epuerta9/gojango/pkg/gojango/admin/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// AdminServiceHandler implements the gRPC AdminService
type AdminServiceHandler struct {
	site      *Site
	bridge    *EntBridge
	entClient interface{} // Generic Ent client interface
}

// NewAdminServiceHandler creates a new admin service handler
func NewAdminServiceHandler(site *Site, bridge *EntBridge) *AdminServiceHandler {
	return &AdminServiceHandler{
		site:      site,
		bridge:    bridge,
		entClient: nil, // Will be set when an Ent client is available
	}
}

// SetEntClient sets the Ent client for database operations
func (h *AdminServiceHandler) SetEntClient(client interface{}) {
	h.entClient = client
}

// ListModels returns all registered models and their configuration
func (h *AdminServiceHandler) ListModels(
	ctx context.Context,
	req *connect.Request[adminpb.ListModelsRequest],
) (*connect.Response[adminpb.ListModelsResponse], error) {
	models := make(map[string]*adminpb.ModelInfo)

	h.site.mu.RLock()
	defer h.site.mu.RUnlock()

	for key, modelAdmin := range h.site.models {
		parts := strings.SplitN(key, ".", 2)
		if len(parts) != 2 {
			continue
		}

		app, modelName := parts[0], parts[1]

		// Convert admin actions
		var actions []*adminpb.AdminAction
		for _, action := range modelAdmin.actions {
			actions = append(actions, &adminpb.AdminAction{
				Name:        action.Name,
				Description: action.Description,
			})
		}

		modelInfo := &adminpb.ModelInfo{
			App:                   app,
			Name:                  modelName,
			VerboseName:          modelAdmin.verboseName,
			VerboseNamePlural:    modelAdmin.verboseNamePlural,
			ListDisplay:          modelAdmin.listDisplay,
			SearchFields:         modelAdmin.searchFields,
			ListFilter:           modelAdmin.listFilter,
			ReadonlyFields:       modelAdmin.readonly,
			Exclude:              modelAdmin.exclude,
			Actions:              actions,
			ListPerPage:          int32(modelAdmin.listPerPage),
			Ordering:             strings.Join(modelAdmin.ordering, ","),
			ShowFullResultCount:  true,
			Permissions: &adminpb.ModelPermissions{
				Add:    true,
				Change: true,
				Delete: true,
				View:   true,
			},
		}

		models[key] = modelInfo
	}

	response := &adminpb.ListModelsResponse{
		Models: models,
		Site: &adminpb.SiteInfo{
			Name:        "admin",
			HeaderTitle: "Gojango Administration",
			IndexTitle:  "Site Administration",
		},
	}

	return connect.NewResponse(response), nil
}

// GetModelSchema returns detailed schema information for a model
func (h *AdminServiceHandler) GetModelSchema(
	ctx context.Context,
	req *connect.Request[adminpb.GetModelSchemaRequest],
) (*connect.Response[adminpb.GetModelSchemaResponse], error) {
	modelKey := fmt.Sprintf("%s.%s", req.Msg.App, req.Msg.Model)

	h.site.mu.RLock()
	modelAdmin, exists := h.site.models[modelKey]
	h.site.mu.RUnlock()

	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("model %s not found", modelKey))
	}

	// Get model info
	modelInfo := &adminpb.ModelInfo{
		App:                  req.Msg.App,
		Name:                 req.Msg.Model,
		VerboseName:         modelAdmin.verboseName,
		VerboseNamePlural:   modelAdmin.verboseNamePlural,
		ListDisplay:         modelAdmin.listDisplay,
		SearchFields:        modelAdmin.searchFields,
		ListFilter:          modelAdmin.listFilter,
		ReadonlyFields:      modelAdmin.readonly,
		Exclude:             modelAdmin.exclude,
		ListPerPage:         int32(modelAdmin.listPerPage),
		Ordering:            strings.Join(modelAdmin.ordering, ","),
		ShowFullResultCount: true,
		Permissions: &adminpb.ModelPermissions{
			Add:    true,
			Change: true,
			Delete: true,
			View:   true,
		},
	}

	// Get field information using reflection
	var fields []*adminpb.FieldInfo
	if modelAdmin.model != nil {
		modelType := reflect.TypeOf(modelAdmin.model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		reflector := &EntModelReflector{modelType: modelType}
		fieldInfos := reflector.GetFields()
		
		for _, fieldInfo := range fieldInfos {
			field := &adminpb.FieldInfo{
				Name:         fieldInfo.Name,
				FieldType:    fieldInfo.FieldType,
				VerboseName:  fieldInfo.VerboseName,
				HelpText:     fieldInfo.HelpText,
				Required:     fieldInfo.Required,
				Editable:     fieldInfo.Editable,
				Blank:        fieldInfo.Blank,
				Null:         fieldInfo.Null,
				MaxLength:    int32(fieldInfo.MaxLength),
				Unique:       fieldInfo.Unique,
				RelatedModel: fieldInfo.RelatedModel,
				WidgetType:   fieldInfo.WidgetType,
			}
			fields = append(fields, field)
		}
	}

	response := &adminpb.GetModelSchemaResponse{
		ModelInfo: modelInfo,
		Fields:    fields,
	}

	return connect.NewResponse(response), nil
}

// ListObjects returns paginated list of model objects
func (h *AdminServiceHandler) ListObjects(
	ctx context.Context,
	req *connect.Request[adminpb.ListObjectsRequest],
) (*connect.Response[adminpb.ListObjectsResponse], error) {
	modelKey := fmt.Sprintf("%s.%s", req.Msg.App, req.Msg.Model)

	h.site.mu.RLock()
	modelAdmin, exists := h.site.models[modelKey]
	h.site.mu.RUnlock()

	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("model %s not found", modelKey))
	}

	// Set default pagination
	page := req.Msg.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.Msg.PageSize
	if pageSize < 1 {
		pageSize = int32(modelAdmin.listPerPage)
	}
	if pageSize > 200 { // Max page size
		pageSize = 200
	}

	// TODO: Implement real Ent database queries when client is available
	var objects []*adminpb.ObjectData
	var totalCount int32
	
	if h.entClient != nil {
		// When Ent client is available, implement real database queries here
		// This would involve:
		// 1. Using reflection to call the appropriate Query() method on the Ent client
		// 2. Applying filters, search, and ordering
		// 3. Handling pagination with Offset() and Limit()
		// 4. Converting Ent objects to protobuf ObjectData using ConvertEntObjectToObjectData
		
		// For now, fall back to mock data
		objects = h.getMockObjects(req.Msg.App, req.Msg.Model, int(page), int(pageSize))
		totalCount = int32(len(objects) * 10)
	} else {
		// Return mock data when no Ent client is available
		objects = h.getMockObjects(req.Msg.App, req.Msg.Model, int(page), int(pageSize))
		totalCount = int32(len(objects) * 10)
	}

	response := &adminpb.ListObjectsResponse{
		Objects:       objects,
		TotalCount:    totalCount,
		Page:          page,
		PageSize:      pageSize,
		HasNext:       page*pageSize < totalCount,
		HasPrevious:   page > 1,
		TotalPages:    (totalCount + pageSize - 1) / pageSize,
		DisplayFields: modelAdmin.listDisplay,
	}

	return connect.NewResponse(response), nil
}

// getMockObjects returns mock data for testing
func (h *AdminServiceHandler) getMockObjects(app, model string, page, pageSize int) []*adminpb.ObjectData {
	var objects []*adminpb.ObjectData

	switch model {
	case "user":
		for i := 0; i < pageSize; i++ {
			id := (page-1)*pageSize + i + 1
			fields := map[string]*structpb.Value{
				"id":       structpb.NewNumberValue(float64(id)),
				"username": structpb.NewStringValue(fmt.Sprintf("user%d", id)),
				"email":    structpb.NewStringValue(fmt.Sprintf("user%d@example.com", id)),
				"is_active": structpb.NewBoolValue(id%2 == 1),
			}
			
			objects = append(objects, &adminpb.ObjectData{
				Id:                strconv.Itoa(id),
				Fields:            fields,
				StrRepresentation: fmt.Sprintf("User: user%d", id),
			})
		}
	case "post":
		for i := 0; i < pageSize; i++ {
			id := (page-1)*pageSize + i + 1
			fields := map[string]*structpb.Value{
				"id":      structpb.NewNumberValue(float64(id)),
				"title":   structpb.NewStringValue(fmt.Sprintf("Post Title %d", id)),
				"status":  structpb.NewStringValue("published"),
				"author_id": structpb.NewNumberValue(1),
			}
			
			objects = append(objects, &adminpb.ObjectData{
				Id:                strconv.Itoa(id),
				Fields:            fields,
				StrRepresentation: fmt.Sprintf("Post: Post Title %d", id),
			})
		}
	case "category":
		for i := 0; i < pageSize; i++ {
			id := (page-1)*pageSize + i + 1
			fields := map[string]*structpb.Value{
				"id":   structpb.NewNumberValue(float64(id)),
				"name": structpb.NewStringValue(fmt.Sprintf("Category %d", id)),
				"slug": structpb.NewStringValue(fmt.Sprintf("category-%d", id)),
			}
			
			objects = append(objects, &adminpb.ObjectData{
				Id:                strconv.Itoa(id),
				Fields:            fields,
				StrRepresentation: fmt.Sprintf("Category: Category %d", id),
			})
		}
	}

	return objects
}

// GetObject returns a single object by ID
func (h *AdminServiceHandler) GetObject(
	ctx context.Context,
	req *connect.Request[adminpb.GetObjectRequest],
) (*connect.Response[adminpb.GetObjectResponse], error) {
	// TODO: Implement get single object
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("GetObject not implemented yet"))
}

// CreateObject creates a new object
func (h *AdminServiceHandler) CreateObject(
	ctx context.Context,
	req *connect.Request[adminpb.CreateObjectRequest],
) (*connect.Response[adminpb.CreateObjectResponse], error) {
	// TODO: Implement create object
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("CreateObject not implemented yet"))
}

// UpdateObject updates an existing object
func (h *AdminServiceHandler) UpdateObject(
	ctx context.Context,
	req *connect.Request[adminpb.UpdateObjectRequest],
) (*connect.Response[adminpb.UpdateObjectResponse], error) {
	// TODO: Implement update object
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("UpdateObject not implemented yet"))
}

// DeleteObject deletes a single object
func (h *AdminServiceHandler) DeleteObject(
	ctx context.Context,
	req *connect.Request[adminpb.DeleteObjectRequest],
) (*connect.Response[adminpb.DeleteObjectResponse], error) {
	// TODO: Implement delete object
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("DeleteObject not implemented yet"))
}

// DeleteObjects deletes multiple objects
func (h *AdminServiceHandler) DeleteObjects(
	ctx context.Context,
	req *connect.Request[adminpb.DeleteObjectsRequest],
) (*connect.Response[adminpb.DeleteObjectsResponse], error) {
	// TODO: Implement bulk delete
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("DeleteObjects not implemented yet"))
}

// ExecuteAction executes a custom admin action
func (h *AdminServiceHandler) ExecuteAction(
	ctx context.Context,
	req *connect.Request[adminpb.ExecuteActionRequest],
) (*connect.Response[adminpb.ExecuteActionResponse], error) {
	// TODO: Implement admin actions
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("ExecuteAction not implemented yet"))
}

// ListActions returns available actions for a model
func (h *AdminServiceHandler) ListActions(
	ctx context.Context,
	req *connect.Request[adminpb.ListActionsRequest],
) (*connect.Response[adminpb.ListActionsResponse], error) {
	modelKey := fmt.Sprintf("%s.%s", req.Msg.App, req.Msg.Model)

	h.site.mu.RLock()
	modelAdmin, exists := h.site.models[modelKey]
	h.site.mu.RUnlock()

	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("model %s not found", modelKey))
	}

	var actions []*adminpb.AdminAction
	for _, action := range modelAdmin.actions {
		actions = append(actions, &adminpb.AdminAction{
			Name:        action.Name,
			Description: action.Description,
		})
	}

	response := &adminpb.ListActionsResponse{
		Actions: actions,
	}

	return connect.NewResponse(response), nil
}

// SearchObjects performs search across model objects
func (h *AdminServiceHandler) SearchObjects(
	ctx context.Context,
	req *connect.Request[adminpb.SearchObjectsRequest],
) (*connect.Response[adminpb.SearchObjectsResponse], error) {
	// TODO: Implement search functionality
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("SearchObjects not implemented yet"))
}