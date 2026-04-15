package handlers

import (
	"regexp"
	"shopTemplate/app/config"
	"shopTemplate/app/db"
	"shopTemplate/app/helpers"
	"shopTemplate/app/models"
	"shopTemplate/app/views/checkout"

	"github.com/anthdm/superkit/event"
	"github.com/anthdm/superkit/kit"
	"github.com/anthdm/superkit/validate"
)

func HandleCheckoutIndex(kit *kit.Kit) error {
	cart := helpers.GetCart(kit)
	cfg := config.Get()
	return RenderWithLayout(kit, checkout.Index(cart, make(map[string]string), make(validate.Errors), cfg))
}

func HandleCheckoutSuccess(kit *kit.Kit) error {
	cfg := config.Get()
	return RenderWithLayout(kit, checkout.Success(cfg))
}

func HandleCheckoutCreate(kit *kit.Kit) error {
	errors := make(validate.Errors)
	phone := kit.Request.FormValue("phone")
	cfg := config.Get()

	// Server-side validation for exactly 8 digits
	matched, _ := regexp.MatchString(`^[0-9]{8}$`, phone)
	if !matched {
		errors.Add("phone", "Phone number must be exactly 8 digits")
	}

	if len(errors) > 0 {
		cart := helpers.GetCart(kit)
		values := map[string]string{
			"firstName": kit.Request.FormValue("firstName"),
			"lastName":  kit.Request.FormValue("lastName"),
			"email":     kit.Request.FormValue("email"),
			"address":   kit.Request.FormValue("address"),
			"city":      kit.Request.FormValue("city"),
			"phone":     phone,
		}
		return RenderWithLayout(kit, checkout.Index(cart, values, errors, cfg))
	}

	// 1. Create the Order
	cart := helpers.GetCart(kit)
	order := models.Order{
		FirstName: kit.Request.FormValue("firstName"),
		LastName:  kit.Request.FormValue("lastName"),
		Email:     kit.Request.FormValue("email"),
		Address:   kit.Request.FormValue("address"),
		City:      kit.Request.FormValue("city"),
		Phone:     phone,
		Total:     calculateTotal(cart),
		Status:    "pending",
	}

	if err := db.Get().Create(&order).Error; err != nil {
		return err
	}

	// 2. Create Order Items
	for _, item := range cart.Items {
		orderItem := models.OrderItem{
			OrderID:      order.ID,
			ProductID:    item.Product.ID,
			ProductName:  item.Product.Name,
			ProductImage: item.Product.Image,
			Quantity:     item.Quantity,
			Price: func() float64 {
				if item.Product.PromotionPrice > 0 {
					return item.Product.PromotionPrice
				}
				return item.Product.Price
			}(),
		}
		db.Get().Create(&orderItem)

		// Optional: Deduct stock here
		// db.Get().Model(&item.Product).Update("stock", item.Product.Stock - item.Quantity)
	}

	// 3. Clear the cart
	sess := kit.GetSession("session")
	sess.Values["cart"] = &models.Cart{Items: make(map[uint]*models.CartItem)}
	sess.Save(kit.Request, kit.Response)

	// Emit the order placement event
	event.Emit("order.placed", order)

	// Redirect to the success page
	kit.Response.Header().Set("HX-Redirect", "/checkout/success")
	return nil
}

func calculateTotal(cart *models.Cart) float64 {
	var total float64
	for _, item := range cart.Items {
		price := item.Product.Price
		if item.Product.PromotionPrice > 0 {
			price = item.Product.PromotionPrice
		}
		total += price * float64(item.Quantity)
	}
	return total
}
