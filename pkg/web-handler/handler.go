package web_handler

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/go-playground/validator/v10"
	"k8s-network-probe/pkg/probe"
	target_store "k8s-network-probe/pkg/target-store"
	"k8s.io/klog/v2"
	"net/http"
)

type APIHandler struct {
	Store target_store.TargetStore
}

// view 处理函数 2个参数  request response
func (api *APIHandler) GetTargetsFromStore(request *restful.Request, response *restful.Response) {
	ts := api.Store.GetTargets()

	localIp := request.QueryParameter("local_ip")
	nodeName := request.QueryParameter("node_name")
	runType := request.QueryParameter("run_type")
	klog.Infof("[GetTargetsFromStore.request.print][runType:%v][nodeName:%v][localIp:%v]", runType, nodeName, localIp)
	//if err != nil {
	//	klog.Errorf("api.Store.Load.err:%v", err)
	//	response.WriteErrorString(http.StatusInternalServerError, err.Error()+"\n")
	//}

	response.WriteHeaderAndEntity(http.StatusOK, ts)
}

func (api *APIHandler) AddTargets(request *restful.Request, response *restful.Response) {

	ts := make([]*probe.ProbeTarget, 0)
	//ts := new(probe.ProbeTargets)
	err := request.ReadEntity(&ts)
	if err != nil {
		klog.Errorf("AddTargets.ReadEntity.err.print:%v", err)

		response.WriteErrorString(400, err.Error())
		return
	}

	//实例化验证器
	// 这里为什么要用validator 因为 go-restful web框架没有自带验证器
	// 如果之间用gin框架 没必要自己再整validator

	validate := validator.New()
	for _, t := range ts {
		t := t
		err = validate.Struct(t)
		if err != nil {
			klog.Errorf("AddTargets.err.print:%v", err)
			response.WriteErrorString(400, err.Error())
			return
		}
	}

	klog.Infof("AddTargets.err.print:%v", err)
	addNum, totalNum, err := api.Store.UpdateTargets(ts)
	if err != nil {
		klog.Errorf("api.Store.UpdateTargets.err:%v", err)
		response.WriteErrorString(500, err.Error())
		return
	}
	addResult := map[string]int{
		"add_num":   addNum,
		"total_num": totalNum,
	}

	response.WriteHeaderAndEntity(http.StatusOK, addResult)
}

func NewHandler(store target_store.TargetStore) (http.Handler, error) {

	aPIHandler := &APIHandler{Store: store}
	wsContainer := restful.NewContainer()
	wsContainer.EnableContentEncoding(true)

	apiV1Ws := new(restful.WebService)

	//InstallFilters(apiV1Ws, cManager)

	apiV1Ws.Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	wsContainer.Add(apiV1Ws)

	apiV1Ws.Route(
		apiV1Ws.GET("/hello").
			To(handleHello))

	apiV1Ws.Route(
		apiV1Ws.GET("/probe-targets").
			To(aPIHandler.GetTargetsFromStore))

	apiV1Ws.Route(
		apiV1Ws.POST("/probe-targets").
			To(aPIHandler.AddTargets))

	return wsContainer, nil
}

func handleHello(request *restful.Request, response *restful.Response) {

	response.WriteHeaderAndEntity(http.StatusOK, "hello")
	return
}
