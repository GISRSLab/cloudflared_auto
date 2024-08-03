package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rivo/tview"
)

// 确认选择
func ConfirmSelection(app *tview.Application, data cloudflare.UnvalidatedIngressRule, list *tview.List, item string, textview *tview.TextView, page *tview.Flex) {

	modal := tview.NewModal().
		SetText(fmt.Sprintf("You selected: \n%s\nDo you want to confirm?", item)).
		AddButtons([]string{"Confirm", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Confirm" {
				app.SetRoot(page, true).SetFocus(list)
				fmt.Printf("Confirmed selection: %s\n", item)
				go func() {
					err := installService(strings.Split(data.Hostname, ".")[0], item, data.Hostname, app, textview)
					if err != nil {
						app.QueueUpdateDraw(func() {
							fmt.Fprintf(textview, "Installation failed: %v\n", err)
						})
					} else {
						app.QueueUpdateDraw(func() {
							fmt.Fprintf(textview, "Service installed and started successfully\n")
						})
					}
				}()
			}
			app.SetRoot(page, true).SetFocus(list)
		})
	app.SetRoot(modal, true).SetFocus(modal)
}

func CreateInstallPage(app *tview.Application, data []cloudflare.UnvalidatedIngressRule) (*tview.Flex, *tview.List, *tview.TextView) {
	// 创建一个列表
	list := tview.NewList().
		AddItem("返回[Back]", "", 'q', func() {
		})

	textview := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	textview.SetBorder(true).SetTitle("Service Installation Log")

	// 设置布局
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true)
	page.SetTitle("--安装服务[Install win service]--")
	page.AddItem(textview, 0, 1, false)
	// 添加数据到列表
	for i, item := range data {
		name := item.Hostname + " ===> " + item.Service
		list.AddItem(name, "", rune('1'+i), func() {
			ConfirmSelection(app, item, list, name, textview, page)
		})
	}

	return page, list, textview
}

func CreateServicePage(app *tview.Application, data []cloudflare.UnvalidatedIngressRule) (*tview.Flex, *tview.List, *tview.TextView) {
	// 创建一个列表
	list := tview.NewList().
		AddItem("返回[Back]", "", 'q', func() {
		})

	textview := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	textview.SetBorder(true).SetTitle("Service List")
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true).
		AddItem(textview, 0, 1, false)
	page.SetTitle("--删除服务[Delete win service]--")
	return page, list, textview
}

func CreateConfigurePage(app *tview.Application, api *cloudflare.API, ctx context.Context, rc *cloudflare.ResourceContainer, tID string, zID string) (*tview.Flex, *tview.Form) {
	textview := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	textview.SetBorder(true).SetTitle("Configure log")

	form := tview.NewForm().
		AddInputField("公网子域名[Public subdomain]-eg. test", "", 20, nil, nil).
		AddInputField("内网地址[Intranet url]-eg. localhost:3389", "", 20, nil, nil)

	form.AddButton("Submit", func() {
		// 在这里处理提交逻辑
		subdomain := strings.Replace(form.GetFormItem(0).(*tview.InputField).GetText(), " ", "", -1)
		url := strings.Replace(form.GetFormItem(1).(*tview.InputField).GetText(), " ", "", -1)

		hostname := subdomain + "." + os.Getenv("SITE")
		tunnelConfig, err1 := api.GetTunnelConfiguration(ctx, rc, tID)
		if err1 != nil {
			fmt.Fprintf(textview, "Failed: %s", err1.Error())
		} else {
			backupIngress := tunnelConfig.Config.Ingress
			params := cloudflare.TunnelConfigurationParams{
				TunnelID: tID,
				Config: cloudflare.TunnelConfiguration{
					Ingress: append(tunnelConfig.Config.Ingress[:len(tunnelConfig.Config.Ingress)-1], cloudflare.UnvalidatedIngressRule{
						Hostname: hostname,
						Path:     "",
						Service:  "rdp://" + url,
					}, cloudflare.UnvalidatedIngressRule{
						Service: "http_status:404",
					}),
				},
			}

			var isproxied bool = true
			dnsRecordParams := cloudflare.CreateDNSRecordParams{
				Content: tID + ".cfargotunnel.com",
				Name:    hostname,
				Type:    "CNAME",
				ID:      zID,
				Proxied: &isproxied,
				Comment: "From API creation.",
			}
			theRc := cloudflare.ZoneIdentifier(zID)
			_, err1 := api.CreateDNSRecord(ctx, theRc, dnsRecordParams)
			result, err2 := api.UpdateTunnelConfiguration(ctx, rc, params)
			if err2 != nil {
				fmt.Fprintf(textview, "Failed: %s", err2.Error())
			} else if err1 != nil {
				//recover configuration
				_, _ = api.UpdateTunnelConfiguration(ctx, rc, cloudflare.TunnelConfigurationParams{
					TunnelID: tID,
					Config: cloudflare.TunnelConfiguration{
						Ingress: backupIngress,
					},
				})
				fmt.Fprintf(textview, "Failed: %s", err1.Error())
			} else {
				fmt.Fprintf(textview, "Success: %v", result.Config.Ingress)
			}
		}
	})

	form.AddButton("Back", func() {})
	form.SetBorder(true).SetTitle("Enter Subdomain and Intranet Address").SetTitleAlign(tview.AlignLeft)
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(textview, 0, 1, false)
	page.SetTitle("--配置公网--")
	return page, form
}

func StartApp(data []cloudflare.UnvalidatedIngressRule, api *cloudflare.API, ctx context.Context, rc *cloudflare.ResourceContainer, tID string, zID string) {
	// 数据
	// data := []string{
	// 	"Item 1",
	// 	"Item 2",
	// 	"Item 3",
	// 	"Item 4",
	// 	"Item 5",
	// }
	app := tview.NewApplication()

	installPage, install_list, install_textview := CreateInstallPage(app, data)
	servicePage, service_list, service_textview := CreateServicePage(app, data)
	configurePage, form := CreateConfigurePage(app, api, ctx, rc, tID, zID)

	mainTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetText("请在本软件同目录放置.env文件\n请使用管理员权限运行本软件\n本软件只提供cloudflare tunnel的public hostname 访问远程rdp及其自动注册为windows服务\n.env文件示例:\n\nCLOUDFLARE_API_KEY=xasdfnlasldfasdf\nCLOUDFLARE_API_EMAIL=xxx@gmail.com\nCLOUDFLARE_TUNNEL_ID=szcva-asdf-vkak3nfald-ansdf\nCLOUDFLARE_ACCOUNT_ID=ncvalsdfasdnfkla\nCLOUDFLARE_ZONE_ID=anvlasdf\nSITE=xxxxx.com\n")

	mainTextView.SetBorder(true).SetTitle("Tips")
	// main page
	mainPage := tview.NewFlex().
		AddItem(tview.NewList().
			AddItem("退出[Quit]", "", 'q', func() {
				app.Stop()
			}).
			AddItem("安装服务[Install win service]", "", '1', func() {
				tunnelConfig, err := api.GetTunnelConfiguration(ctx, rc, tID)
				if err != nil {
					log.Fatal(err)
				}
				// Print user details
				var newdata []cloudflare.UnvalidatedIngressRule
				for _, item := range tunnelConfig.Config.Ingress {
					if strings.HasPrefix(item.Service, "rdp") {
						newdata = append(newdata, item)
					}
				}

				install_list.Clear().AddItem("返回[Back]", "", 'q', func() {})
				for i, item := range newdata {
					name := item.Hostname + " ===> " + item.Service
					install_list.AddItem(name, "", rune('1'+i), func() {
						ConfirmSelection(app, item, install_list, name, install_textview, installPage)
					})
				}

				app.SetRoot(installPage, true).SetFocus(installPage)
			}).
			AddItem("删除服务[Delete win service]", "", '2', func() {
				list_service, err := listService()
				if err != nil {
					go func() {
						app.QueueUpdateDraw(func() {
							fmt.Fprint(service_textview, "\n"+err.Error())
						})
					}()
				} else {
					service_list.Clear().AddItem("返回[Back]", "", 'q', func() {})
					for i, service := range list_service {
						service_list.AddItem(service, "", rune('1'+i), func() {
							err := deleteService(service)
							if err != nil {
								go func() {
									app.QueueUpdateDraw(func() {
										fmt.Fprintf(service_textview, "\nFailed: %s", err.Error())
									})
								}()
							} else {
								service_list.RemoveItem(i + 1)
								go func() {
									app.QueueUpdateDraw(func() {
										fmt.Fprintf(service_textview, "\n%s has been deleted!", service)
									})
								}()
							}
						})
					}
				}

				app.SetRoot(servicePage, true).SetFocus(servicePage)
			}).
			AddItem("公网转发[Configure public host]", "", '3', func() {
				app.SetRoot(configurePage, true).SetFocus(configurePage)
			}),
			0, 1, true).
		AddItem(mainTextView, 0, 1, false)
	install_list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index == 0 {
			app.SetRoot(mainPage, true).SetFocus(mainPage)
		}
	})
	service_list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index == 0 {
			app.SetRoot(mainPage, true).SetFocus(mainPage)
		}
	})

	form.GetButton(1).SetSelectedFunc(func() {
		app.SetRoot(mainPage, true).SetFocus(mainPage)
	})
	// 设置应用的根页面
	app.SetRoot(mainPage, true).SetFocus(mainPage)

	// 运行应用
	if err := app.Run(); err != nil {
		panic(err)
	}
}
