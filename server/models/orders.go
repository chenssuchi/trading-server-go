package models

type Orders struct {
	Id       int    `json:"id" gorm:"column:id"`
	OrderId  int64  `json:"order_id" gorm:"column:order_id"`
	Seller   string `json:"seller" gorm:"column:seller"`
	Buyer    string `json:"buyer" gorm:"column:buyer"`
	Price    string `json:"price" gorm:"column:price"`
	Quantity int64  `json:"quantity" gorm:"column:quantity"`
	SaleAt   string `json:"sale_at" gorm:"column:sale_at"`
	BuyAt    string `json:"buy_at" gorm:"column:buy_at"`
	//SaleAt int64  `json:"sale_at" gorm:"column:sale_at"`
	//BuyAt  int64  `json:"buy_at" gorm:"column:buy_at"`
	IsDeal bool `json:"is_deal" gorm:"column:is_deal"`
}
