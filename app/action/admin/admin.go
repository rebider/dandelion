package admin

import (
	"github.com/nbvghost/dandelion/app/action/admin/activityAction"
	"github.com/nbvghost/dandelion/app/action/admin/companyAction"
	"github.com/nbvghost/dandelion/app/action/admin/contentAction"
	"github.com/nbvghost/dandelion/app/play"
	"github.com/nbvghost/dandelion/app/result"
	"github.com/nbvghost/dandelion/app/service/activity"
	"github.com/nbvghost/dandelion/app/service/admin"
	"github.com/nbvghost/dandelion/app/service/configuration"
	"github.com/nbvghost/dandelion/app/service/content"
	"github.com/nbvghost/dandelion/app/service/express"
	"github.com/nbvghost/dandelion/app/service/goods"
	"github.com/nbvghost/dandelion/app/service/order"
	"github.com/nbvghost/dandelion/app/service/user"
	"github.com/nbvghost/dandelion/app/util"
	"net/url"
	"strings"

	"github.com/nbvghost/glog"

	"github.com/nbvghost/dandelion/app/service/dao"

	"errors"
	"strconv"

	"github.com/nbvghost/gweb"
)

type InterceptorManager struct {
}

func (controller InterceptorManager) ActionBefore(context *gweb.Context) (bool, gweb.Result) {
	//util.Trace(context.Session,"context.Session")
	if context.Session.Attributes.Get(play.SessionAdmin) == nil {
		//http.SetCookie(context.Response, &http.Cookie{Name: "UID", MaxAge:-1, Path: "/"})
		//fmt.Println(context.Request.URL.Path)
		//fmt.Println(context.Request.URL.Query().Encode())
		redirect := ""
		if len(context.Request.URL.Query().Encode()) == 0 {
			redirect = context.Request.URL.Path
		} else {
			redirect = context.Request.URL.Path + "?" + context.Request.URL.Query().Encode()
		}

		//fmt.Println(url.QueryEscape(redirect))
		//http.Redirect(context.Response, context.Request, "/account/loginAdminPage?redirect="+url.QueryEscape(redirect), http.StatusFound)
		return false, &gweb.RedirectToUrlResult{Url: "/account/loginAdminPage?redirect=" + url.QueryEscape(redirect)}
	} else {
		return true, nil
	}
}
func (controller InterceptorManager) ActionBeforeServiceName(context *gweb.Context) string {
	return ""
}
func (controller InterceptorManager) ActionAfter(context *gweb.Context, result gweb.Result) gweb.Result {
	return nil
}

type Controller struct {
	gweb.BaseController
	Admin admin.AdminService
	Goods goods.GoodsService

	ExpressTemplate express.ExpressTemplateService
	Voucher         activity.VoucherService
	//TimeSell        activity.TimeSellService
	Orders        order.OrdersService
	ScoreGoods    activity.ScoreGoodsService
	Configuration configuration.ConfigurationService
	GiveVoucher   activity.GiveVoucherService
	User          user.UserService
	CardItem      activity.CardItemService
	Content       content.ContentService
	Collage       activity.CollageService
}

func (controller *Controller) Init() {
	controller.Interceptors.Set(&InterceptorManager{})
	controller.AddHandler(gweb.ALLMethod("index", controller.indexPage))
	controller.AddHandler(gweb.ALLMethod("goods", controller.GoodsAction))

	controller.AddHandler(gweb.POSMethod("express_template/save", controller.saveExpressTemplate))
	controller.AddHandler(gweb.DELMethod("express_template/{ID}", controller.deleteExpressTemplate))
	controller.AddHandler(gweb.GETMethod("express_template/{ID}", controller.getExpressTemplate))
	controller.AddHandler(gweb.POSMethod("express_template/datatables/list", controller.datatablesExpressTemplate))
	controller.AddHandler(gweb.GETMethod("express_template/list", controller.listExpressTemplate))
	controller.AddHandler(gweb.POSMethod("store_journal/list", controller.storeJournalListAction))
	controller.AddHandler(gweb.POSMethod("order/list", controller.listOrderAction))
	controller.AddHandler(gweb.PUTMethod("order/change", controller.orderChangeAction))
	controller.AddHandler(gweb.POSMethod("configuration/list", controller.configurationListAction))
	controller.AddHandler(gweb.POSMethod("configuration/change", controller.configurationChangeAction))
	controller.AddHandler(gweb.POSMethod("carditem/list", controller.carditemListAction))
	//去掉功能
	//controller.AddHandler(gweb.DELMethod("give_voucher/:GiveVoucherID", controller.giveVoucherDeleteAction))
	controller.AddHandler(gweb.POSMethod("situation", controller.situationAction))
	controller.AddHandler(gweb.POSMethod("admin", controller.Admin.AddItem))
	controller.AddHandler(gweb.GETMethod("admin/{ID}", controller.Admin.GetItem))
	controller.AddHandler(gweb.POSMethod("admin/list", controller.Admin.ListItem))
	controller.AddHandler(gweb.DELMethod("admin/{ID}", controller.Admin.DeleteItem))
	controller.AddHandler(gweb.PUTMethod("admin/{ID}", controller.Admin.ChangePassWork))
	controller.AddHandler(gweb.PUTMethod("admin/authority/{ID}", controller.Admin.ChangeAuthority))
	controller.AddHandler(gweb.ALLMethod("loginOut", controller.loginOutAction))
	//--------------content------------------

	store := &StoreController{}
	controller.AddSubController("/store/", store)

	content := &contentAction.Controller{}
	controller.AddSubController("/content/", content)

	company := &companyAction.Controller{}
	controller.AddSubController("/company/", company)

	timesell := &activityAction.TimesellController{}
	controller.AddSubController("/timesell/", timesell)

	collage := &activityAction.CollageController{}
	controller.AddSubController("/collage/", collage)

	scoreGoods := &activityAction.ScoreGoodsController{}
	controller.AddSubController("/score_goods/", scoreGoods)

	voucher := &activityAction.VoucherController{}
	controller.AddSubController("/voucher/", voucher)

	fullcut := &activityAction.FullcutController{}
	controller.AddSubController("/fullcut/", fullcut)

}

func (controller *Controller) GoodsAction(context *gweb.Context) gweb.Result {
	company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)
	action := context.Request.URL.Query().Get("action")
	Orm := dao.Orm()
	switch action {
	case "del_goods":
		ID, _ := strconv.ParseUint(context.Request.URL.Query().Get("ID"), 10, 64)
		return &gweb.JsonResult{Data: controller.Goods.DeleteGoods(ID)}
	case "list_specification":
		GoodsID, _ := strconv.ParseUint(context.Request.URL.Query().Get("GoodsID"), 10, 64)
		var gts []dao.Specification
		err := controller.Goods.FindWhere(Orm, &gts, company.ID, dao.Specification{GoodsID: GoodsID})
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "OK", gts)}
	case "delete_specification":
		ID, _ := strconv.ParseUint(context.Request.URL.Query().Get("ID"), 10, 64)
		err := controller.Goods.DeleteSpecification(ID)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "删除成功", nil)}
	case "get_goods":
		ID, _ := strconv.ParseUint(context.Request.URL.Query().Get("ID"), 10, 64)
		goodsInfo := controller.Goods.GetGoods(Orm, ID)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(nil, "OK", goodsInfo)}
	case "change_goods":
		context.Request.ParseForm()
		goods_str := context.Request.FormValue("goods")
		specifications_str := context.Request.FormValue("specifications")

		var specifications []dao.Specification
		var item dao.Goods
		err := util.JSONToStruct(goods_str, &item)
		glog.Trace(err)

		var gps []dao.GoodsParams
		err = util.JSONToStruct(item.Params, &gps)
		glog.Trace(err)

		item.Params = util.StructToJSON(&gps)

		var videos []string
		err = util.JSONToStruct(item.Videos, &videos)
		glog.Trace(err)
		item.Videos = util.StructToJSON(&videos)

		var pictures []string
		err = util.JSONToStruct(item.Pictures, &pictures)
		glog.Trace(err)
		item.Pictures = util.StructToJSON(&pictures)

		var images []string
		err = util.JSONToStruct(item.Images, &images)
		glog.Trace(err)
		item.Images = util.StructToJSON(&images)

		err = util.JSONToStruct(specifications_str, &specifications)
		glog.Trace(err)

		err = controller.Goods.SaveGoods(item, specifications)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "修改成功", nil)}

	case "add_goods":
		context.Request.ParseForm()
		goodsStr := context.Request.FormValue("goods")
		specificationsStr := context.Request.FormValue("specifications")

		var specifications []dao.Specification
		var item dao.Goods
		err := util.JSONToStruct(goodsStr, &item)
		glog.Trace(err)

		var gps []dao.GoodsParams
		err = util.JSONToStruct(item.Params, &gps)
		glog.Trace(err)

		item.Params = util.StructToJSON(&gps)

		var videos []string
		err = util.JSONToStruct(item.Videos, &videos)
		glog.Trace(err)
		item.Videos = util.StructToJSON(&videos)

		var pictures []string
		err = util.JSONToStruct(item.Pictures, &pictures)
		glog.Trace(err)
		item.Pictures = util.StructToJSON(&pictures)

		var images []string
		err = util.JSONToStruct(item.Images, &images)
		glog.Trace(err)
		item.Images = util.StructToJSON(&images)

		err = util.JSONToStruct(specificationsStr, &specifications)
		glog.Trace(err)

		item.OID = company.ID
		err = controller.Goods.SaveGoods(item, specifications)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "添加成功", nil)}
	case "activity_goods":
		//company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)
		//Hash := context.Request.URL.Query().Get("Hash")
		dts := &dao.Datatables{}
		//dts.Draw = 10
		//dts.Length = play.Paging
		util.RequestBodyToJSON(context.Request.Body, dts)
		var TimeSellGoodsIDs []uint64
		Orm.Model(&dao.TimeSellGoods{}).Where("OID=?", company.ID).Pluck("GoodsID", &TimeSellGoodsIDs)
		var CollageGoodsIDs []uint64
		Orm.Model(&dao.CollageGoods{}).Where("OID=?", company.ID).Pluck("GoodsID", &CollageGoodsIDs)
		ActivityGoods := make([]uint64, 0)
		ActivityGoods = append(ActivityGoods, TimeSellGoodsIDs...)
		ActivityGoods = append(ActivityGoods, CollageGoodsIDs...)
		dts.NotIDs = ActivityGoods
		draw, recordsTotal, recordsFiltered, list := controller.Goods.DatatablesListOrder(Orm, dts, &[]dao.Goods{}, company.ID, "")
		return &gweb.JsonResult{Data: map[string]interface{}{"data": list, "draw": draw, "recordsTotal": recordsTotal, "recordsFiltered": recordsFiltered}}
	case "list_goods":
		dts := &dao.Datatables{}
		util.RequestBodyToJSON(context.Request.Body, dts)
		draw, recordsTotal, recordsFiltered, list := controller.Goods.DatatablesListOrder(Orm, dts, &[]dao.Goods{}, company.ID, "")
		return &gweb.JsonResult{Data: map[string]interface{}{"data": list, "draw": draw, "recordsTotal": recordsTotal, "recordsFiltered": recordsFiltered}}
	case "get_goods_type_child":
		ID, _ := strconv.ParseUint(context.Request.URL.Query().Get("ID"), 10, 64)
		var goods dao.GoodsTypeChild
		controller.Goods.Get(Orm, ID, &goods)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(nil, "OK", goods)}
	case "list_goods_type_child":
		var gts []dao.GoodsTypeChild
		controller.Goods.FindAll(Orm, &gts)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(nil, "OK", gts)}
	case "list_goods_type_child_id":
		ID, _ := strconv.ParseUint(context.Request.URL.Query().Get("ID"), 10, 64)
		gts := controller.Goods.ListAllGoodsTypeChild(ID)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(nil, "OK", gts)}
	case "list_goods_type_all":
		gts := controller.Goods.ListGoodsTypeForAdmin(company.ID)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(nil, "OK", gts)}
	case "list_goods_type":
		dts := &dao.Datatables{}
		util.RequestBodyToJSON(context.Request.Body, dts)
		draw, recordsTotal, recordsFiltered, list := controller.Goods.DatatablesListOrder(Orm, dts, &[]dao.GoodsType{}, company.ID, "")
		return &gweb.JsonResult{Data: map[string]interface{}{"data": list, "draw": draw, "recordsTotal": recordsTotal, "recordsFiltered": recordsFiltered}}

	//--------------------------------------
	case "del_goods_type":
		ID, _ := strconv.ParseUint(context.Request.URL.Query().Get("ID"), 10, 64)
		return &gweb.JsonResult{Data: controller.Goods.DeleteGoodsType(ID)}
	case "add_goods_type":
		item := &dao.GoodsType{}
		item.OID = company.ID
		err := util.RequestBodyToJSON(context.Request.Body, item)
		glog.Trace(err)
		err = controller.Goods.Add(Orm, item)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "添加成功", nil)}
	case "change_goods_type":
		item := &dao.GoodsType{}
		err := util.RequestBodyToJSON(context.Request.Body, item)
		glog.Trace(err)
		err = controller.Goods.ChangeModel(Orm, item.ID, &dao.GoodsType{Name: item.Name})
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "修改成功", nil)}

	case "del_goods_type_child":
		ID, _ := strconv.ParseUint(context.Request.URL.Query().Get("ID"), 10, 64)
		return &gweb.JsonResult{Data: controller.Goods.DeleteGoodsTypeChild(ID)}
	case "add_goods_type_child":
		item := &dao.GoodsTypeChild{}
		err := util.RequestBodyToJSON(context.Request.Body, item)
		glog.Trace(err)
		err = controller.Goods.Add(Orm, item)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "添加成功", nil)}
	case "change_goods_type_child":
		item := &dao.GoodsTypeChild{}
		err := util.RequestBodyToJSON(context.Request.Body, item)
		glog.Trace(err)
		err = controller.Goods.ChangeModel(Orm, item.ID, &dao.GoodsTypeChild{Name: item.Name, Image: item.Image})
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "修改成功", nil)}

	}

	return &gweb.JsonResult{}
}
func (controller *Controller) carditemListAction(context *gweb.Context) gweb.Result {
	//company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)
	Orm := dao.Orm()
	dts := &dao.Datatables{}
	util.RequestBodyToJSON(context.Request.Body, dts)
	draw, recordsTotal, recordsFiltered, list := controller.CardItem.DatatablesListOrder(Orm, dts, &[]dao.CardItem{}, 0, "")
	return &gweb.JsonResult{Data: map[string]interface{}{"data": list, "draw": draw, "recordsTotal": recordsTotal, "recordsFiltered": recordsFiltered}}
}
func (controller *Controller) situationAction(context *gweb.Context) gweb.Result {
	context.Request.ParseForm()
	StartTime, _ := strconv.ParseInt(context.Request.FormValue("StartTime"), 10, 64)
	EndTime, _ := strconv.ParseInt(context.Request.FormValue("EndTime"), 10, 64)

	results := make(map[string]interface{})

	results["Orders"] = controller.Orders.Situation(StartTime, EndTime)
	results["Voucher"] = controller.Voucher.Situation(StartTime, EndTime)
	results["ScoreGoods"] = controller.ScoreGoods.Situation(StartTime, EndTime)
	results["User"] = controller.User.Situation(StartTime, EndTime)

	return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(nil, "OK", results)}
}

/*func (controller *Controller) giveVoucherDeleteAction(context *gweb.Context) gweb.Result {

	GiveVoucherID, _ := strconv.ParseUint(context.PathParams["GiveVoucherID"], 10, 64)

	err := controller.Rank.Delete(dao.Orm(), &dao.GiveVoucher{}, GiveVoucherID)

	return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "删除成功", nil)}
}*/

func (controller *Controller) configurationChangeAction(context *gweb.Context) gweb.Result {
	company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)
	item := dao.Configuration{}
	util.RequestBodyToJSON(context.Request.Body, &item)
	err := controller.Configuration.ChangeConfiguration(company.ID, item.K, item.V)
	return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "修改成功", nil)}
}
func (controller *Controller) configurationListAction(context *gweb.Context) gweb.Result {
	company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)
	var ks []uint64
	util.RequestBodyToJSON(context.Request.Body, &ks)
	list := controller.Configuration.GetConfigurations(company.ID, ks)
	return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(nil, "OK", list)}
}

func (controller *Controller) orderChangeAction(context *gweb.Context) gweb.Result {
	//company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)
	Orm := dao.Orm()
	context.Request.ParseForm()
	Action := context.Request.FormValue("Action")
	switch Action {
	case "RefundComplete":
		OrdersGoodsID, _ := strconv.ParseUint(context.Request.FormValue("OrdersGoodsID"), 10, 64)
		RefundType, _ := strconv.ParseUint(context.Request.FormValue("RefundType"), 10, 64)
		err, info := controller.Orders.RefundComplete(OrdersGoodsID, RefundType)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, info, nil)}
	case "RefundOk":
		OrdersGoodsID, _ := strconv.ParseUint(context.Request.FormValue("OrdersGoodsID"), 10, 64)
		err, info := controller.Orders.RefundOk(OrdersGoodsID)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, info, nil)}
	case "RefundNo":
		OrdersGoodsID, _ := strconv.ParseUint(context.Request.FormValue("OrdersGoodsID"), 10, 64)
		err, info := controller.Orders.RefundNo(OrdersGoodsID)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, info, nil)}
	case "Cancel":
		ID, _ := strconv.ParseUint(context.Request.FormValue("ID"), 10, 64)
		err, info := controller.Orders.Cancel(ID)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, info, nil)}
	case "CancelOk":
		ID, _ := strconv.ParseUint(context.Request.FormValue("ID"), 10, 64)
		RefundType, _ := strconv.ParseUint(context.Request.FormValue("RefundType"), 10, 64) //退款资金来源	 0=未结算资金退款,1=可用余额退款
		err, info := controller.Orders.CancelOk(ID, RefundType)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, info, nil)}
	case "Deliver":
		ShipName := context.Request.FormValue("ShipName")
		ShipNo := context.Request.FormValue("ShipNo")
		ID, _ := strconv.ParseUint(context.Request.FormValue("ID"), 10, 64)

		err := controller.Orders.Deliver(ShipName, ShipNo, ID)
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "发货成功", nil)}
	case "PayMoney":
		PayMoney, _ := strconv.ParseFloat(context.Request.FormValue("PayMoney"), 64)
		ID, _ := strconv.ParseUint(context.Request.FormValue("ID"), 10, 64)
		err := controller.Orders.ChangeMap(Orm, ID, &dao.Orders{}, map[string]interface{}{"PayMoney": uint64(PayMoney * 100)})
		return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "修改成功", nil)}
		success, message := controller.Orders.ChangeOrdersPayMoney(PayMoney, ID)
		return &gweb.JsonResult{Data: &result.ActionResult{Code: success, Message: message, Data: nil}}

	}

	return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(errors.New("999"), "OK", nil)}

}
func (controller *Controller) storeJournalListAction(context *gweb.Context) gweb.Result {
	//company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)
	Orm := dao.Orm()
	dts := &dao.Datatables{}
	util.RequestBodyToJSON(context.Request.Body, dts)
	draw, recordsTotal, recordsFiltered, list := controller.CardItem.DatatablesListOrder(Orm, dts, &[]dao.StoreJournal{}, 0, "")
	return &gweb.JsonResult{Data: map[string]interface{}{"data": list, "draw": draw, "recordsTotal": recordsTotal, "recordsFiltered": recordsFiltered}}

}
func (controller *Controller) listOrderAction(context *gweb.Context) gweb.Result {

	company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)

	//Orm := dao.Orm()
	dts := &dao.Datatables{}
	util.RequestBodyToJSON(context.Request.Body, dts)

	UserID, _ := strconv.ParseUint(dts.Columns[0].Search.Value, 10, 64)
	PostType, _ := strconv.ParseInt(dts.Columns[1].Search.Value, 10, 64)
	Status := dts.Columns[2].Search.Value

	var StatusList []string
	if !strings.EqualFold(Status, "") {
		StatusList = strings.Split(Status, ",")
	}
	//fmt.Println(dts)
	list, recordsTotal := controller.Orders.ListOrders(UserID, company.ID, int(PostType), StatusList, dts.Length, dts.Start)
	return &gweb.JsonResult{Data: map[string]interface{}{"data": list, "draw": dts.Draw, "recordsTotal": recordsTotal, "recordsFiltered": recordsTotal}}
}

func (controller *Controller) getExpressTemplate(context *gweb.Context) gweb.Result {
	Orm := dao.Orm()
	ID, _ := strconv.ParseUint(context.PathParams["ID"], 10, 64)
	var item dao.ExpressTemplate
	err := controller.ExpressTemplate.Get(Orm, ID, &item)
	return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "", item)}
	//2002
}
func (controller *Controller) deleteExpressTemplate(context *gweb.Context) gweb.Result {
	Orm := dao.Orm()
	ID, _ := strconv.ParseUint(context.PathParams["ID"], 10, 64)

	err := controller.ExpressTemplate.Delete(Orm, &dao.ExpressTemplate{}, ID)

	return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "删除成功", nil)}
}
func (controller *Controller) saveExpressTemplate(context *gweb.Context) gweb.Result {
	company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)
	item := &dao.ExpressTemplate{}
	err := util.RequestBodyToJSON(context.Request.Body, item)
	glog.Trace(err)
	item.OID = company.ID
	err = controller.ExpressTemplate.SaveExpressTemplate(item)
	return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "保存成功", nil)}
}
func (controller *Controller) listExpressTemplate(context *gweb.Context) gweb.Result {
	company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)
	Orm := dao.Orm()
	var list []dao.ExpressTemplate
	err := controller.ExpressTemplate.FindAllByOID(Orm, &list, company.ID)
	return &gweb.JsonResult{Data: (&result.ActionResult{}).SmartError(err, "OK", list)}
}
func (controller *Controller) datatablesExpressTemplate(context *gweb.Context) gweb.Result {
	company := context.Session.Attributes.Get(play.SessionOrganization).(*dao.Organization)
	Orm := dao.Orm()
	dts := &dao.Datatables{}
	util.RequestBodyToJSON(context.Request.Body, dts)
	draw, recordsTotal, recordsFiltered, list := controller.ExpressTemplate.DatatablesListOrder(Orm, dts, &[]dao.ExpressTemplate{}, company.ID, "")
	return &gweb.JsonResult{Data: map[string]interface{}{"data": list, "draw": draw, "recordsTotal": recordsTotal, "recordsFiltered": recordsFiltered}}

}

func (controller *Controller) loginOutAction(context *gweb.Context) gweb.Result {
	context.Session.Attributes.Delete(play.SessionAdmin)
	return &gweb.RedirectToUrlResult{Url: "/admin"}
}

func (controller *Controller) indexPage(context *gweb.Context) gweb.Result {

	return &gweb.HTMLResult{}
}

func (controller *Controller) defaultPage(context *gweb.Context) gweb.Result {

	return &gweb.HTMLResult{}
}
func (controller *Controller) rootPage(context *gweb.Context) gweb.Result {

	return &gweb.RedirectToUrlResult{"index"}
	//return &gweb.HTMLResult{Name: "admin/index.html"}
}
