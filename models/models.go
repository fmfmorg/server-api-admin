package models

type Specification struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductMainType struct {
	Name     string          `json:"name"`
	Subtypes []Specification `json:"subtypes"`
}

type ProductDiscount struct {
	ID      int    `json:"id"`
	Product string `json:"productID"`
	Amount  int    `json:"amount"`
	StartDT int64  `json:"startDT"`
	EndDT   int64  `json:"endDT"`
}

type StockQuantity struct {
	Name     string `json:"name"`
	Address  int    `json:"addressID"`
	Quantity int    `json:"quantity"`
}

type ProductDetails struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description,omitempty"`
	MaterialID      int               `json:"materialID"`
	MetalColorID    int               `json:"metalColorID"`
	ProductTypeID   int               `json:"productTypeID"`
	Price           int               `json:"price"`
	Discounts       []ProductDiscount `json:"discounts"`
	URL             string            `json:"url"`
	PublicImages    []string          `json:"publicImages"`
	AdminImages     []string          `json:"adminImages"`
	StockQuantities []StockQuantity   `json:"stockQuantities"`
	TotalSales      int               `json:"totalSales,omitempty"`
	CreatedAt       int64             `json:"createdAt,omitempty"`
	IsRetired       bool              `json:"isRetired"`
}

type OrderOverviewItem struct {
	OrderID                    int   `json:"orderID"`
	OrderStatusID              int   `json:"orderStatusID"`
	OrderDate                  int64 `json:"orderDate"`
	TotalOrderAmountExDelivery int   `json:"totalOrderAmountExDelivery"`
	DeliveryMethodID           int   `json:"deliveryMethodID"`
}

type DeliveryMethod struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Cost            int    `json:"cost"`
	MinSpendForFree int    `json:"minSpendForFree"`
	RegionName      string `json:"regionName"`
}
