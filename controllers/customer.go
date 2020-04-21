package controllers

import (
	. "customer_managenment/api"
	"customer_managenment/models"
	. "customer_managenment/tool"
	"encoding/json"
	"github.com/astaxie/beego"
)

type CustomerController struct {
	beego.Controller
}


// 解析参数和校验参数,post
func (c *CustomerController) AnalysisAndVerify(request Verify) bool {
	if c.Ctx.Request.Method != "POST" {
		return false
	}

	// 参数解析失败
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, request); err != nil {
		Logs.Info("error parsing parameters")
		return false
	}

	// 校验参数失败
	if !request.VerifyInputPara() {
		Logs.Info("input para error")
		return false
	}
	return true
}


// NewCustomer 创建客户
func (c *CustomerController)CreateCustomer() {
	// 实例新增返回体，并默认初始值
	response := new(ResponseNewCustomer)
	response.Code = 1
	response.Msg = "success"
	request := new(RequestNewCustomer)
	// 解析校验输入的参数
	if res := c.AnalysisAndVerify(request); res {
		// 判断该 OpenApiToken 是否已有注册
		if  models.JudgeIsExists("UtCustomer", "OpenApiToken", request.OpenApiToken) &&
			request.OpenApiToken != "" {
			Logs.Info(request.OpenApiToken, "already exist, don't register again")
			response.Msg = "the customer is registered"
		} else {
			// 数据库新增客户
			if err := models.InsertCustomer(&request.UtCustomer); err != nil {
				Logs.Error(request.OpenApiToken, "register fail, because", err)
				response.Msg = "failed to register customer"
			} else {
				MapCustomerDetail(response, &request.UtCustomer)
			}
		}
	} else {
		response.Msg = "incoming parameter error"
	}
	c.Data["json"] = response
	c.ServeJSON()
	return
}


// DeleteCustomer 删除客户
func (c *CustomerController)DeleteCustomer() {
	response := new(ResponseDelCustomer)
	response.Code = 1
	response.Msg = "success"
	request := new(RequestDelCustomer)
	// 解析校验输入的参数
	if res := c.AnalysisAndVerify(request); res {
		// 通过apiToken删除客户，通过id删除客户
		// 如果id存在，则通过id删除客户
		// 如果apiToken存在，则通过token删除客户
		if  !models.JudgeIsExists("UtCustomer", "Id", request.Id) || (request.OpenApiToken != "" &&
			!models.JudgeIsExists("UtCustomer", "OpenApiToken", request.OpenApiToken)) {
			Logs.Info("customer", request.Id, request.OpenApiToken, "not exist")
			response.Msg = "deleted customer does not exist"
		} else {
			if !models.RemoveCustomer(request.Id, request.OpenApiToken) {
				Logs.Error("delete customer", request.Id, request.OpenApiToken, "fail")
				response.Msg = "failed to delete customer"
			} else {
				Logs.Info("delete", request.Id, request.OpenApiToken, "success")
				response.Code = 0
			}
		}
	} else {
		response.Msg = "incoming parameter error"
	}
	response.Data.OpenApiToken = request.OpenApiToken
	response.Data.Id = request.Id
	c.Data["json"] = response
	c.ServeJSON()
	return
}

// CreateCustomerFollow新增客户跟进
func (c *CustomerController)CreateCustomerFollow() {
	request := new(RequestNewFollow)
	response := new(ResponseNewFollow)
	response.Code = 1
	response.Msg = "success"
	// 解析校验输入的参数
	if res := c.AnalysisAndVerify(request); res {
		// TODO 判断的权限以及合法性
		//判断是否存在customer
		if customer := models.GetCustomerById(request.CustomerId); customer == nil {
			Logs.Info("ut_customer_id not exist")
			response.Msg = "the customer to be followed up dose not exist"
		} else {
			request.Customer = customer
			if err := models.InsertCustomerFollow(&request.CustomerFollowUp); err != nil {
				// 新数据插入数据库失败
				Logs.Error("new customer follow up failed, because", err)
				response.Msg = "new customer follow up failed"
			} else {
				// 添加成功，更新返回题
				Logs.Info("new follow up success")
				response.Code = 0
				MapCustomerFollow(&response.CustomerFollowUp, &request.CustomerFollowUp)
			}
		}
	} else {
		response.Msg = "incoming parameter error"
	}
	c.Data["json"] = response
	c.ServeJSON()
	return
}

// DeleteCustomerFollow删除客户跟进
func (c *CustomerController)DeleteCustomerFollow() {
	request := new(RequestDelFollow)
	response := new(ResponseDelFollow)
	response.Code = 1
	response.Msg = "success"
	// 解析校验输入的参数
	if res := c.AnalysisAndVerify(request); res {
		// 通过跟进的id判断跟进是否存在
		if !models.JudgeIsExists("CustomerFollowUp", "Id", request.Id) {
			Logs.Info("customer follow up", request.Id, "not exist")
			response.Msg = "deleted customer follow up does not exist"
		} else {
			custerId, res := models.RemoveCustomerFollow(request.Id)
			if !res {
				Logs.Error("delete customer follow up", request.Id, "fail")
				response.Msg = "failed to delete customer follow up"
			} else {
				Logs.Info("delete", request.Id, "success")
				response.Code = 0
				response.Data.Id = request.Id
				response.Data.CustomerId = custerId
			}
		}
	} else {
		response.Msg = "incoming parameter error"
	}
	c.Data["json"] = response
	c.ServeJSON()
	return
}

// ShowCustomerDetail 展示客户详情，包括跟进，变更
func (c *CustomerController)ShowCustomerDetail()  {
	request := new(RequestShowCustomer)
	response := new(ResponseShowCustomer)
	response.Code = 1
	response.Msg = "success"
	if res, err := c.GetInt("customer_id"); err == nil {
		request.CustomerId = res
		if res := request.VerifyInputPara(); res {
			if res := models.GetCustomerById(request.CustomerId); res == nil {
				Logs.Info("customer dose not exist, customer_id:", request.CustomerId)
				response.Msg = "customer dose not exist"
			} else {
				response.Code = 0
				MapCustomerDetail(&response.ResponseNewCustomer, res)
				//response.UtCustomer = *res
			}
		} else {
			response.Msg = "incoming parameter error"
		}
	} else {
		response.Msg = "incoming parameter error"
	}
	c.Data["json"] = response
	c.ServeJSON()
	return
}

