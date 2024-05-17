package bootlrLogs

import (
	"context"
	"encore.dev/rlog"
)

type Response struct {
	Res string `json:"res"`
}

//encore:api public method=GET path=/bootlr-log-visitor
func BootlrLogVisitor(ctx context.Context) (Response, error){
	rlog.Info("VISIT LOG", "==> ", "NEW VISITOR")
	
	response := Response{
		Res: "",
	}
	return response, nil
}

//encore:api public method=GET path=/bootlr-log-product-click
func BootlrLogProductClick(ctx context.Context) (Response, error){
	rlog.Info("CLICK LOG", "==> ", "PRODUCT CLICKED")
	
	response := Response{
		Res: "",
	}
	return response, nil
}