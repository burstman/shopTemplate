package helpers

import (
	"shopTemplate/app/models"
)

// CalculateItemPrice calculates the total price for a specific cart item, applying bundle discounts if applicable.
// It prioritizes product-specific bundles if they exist, otherwise falls back to global bundles.
func CalculateItemPrice(item *models.CartItem, globalBundles []models.Bundle) models.Currency {
	price := item.Product.Price
	if item.Product.PromotionPrice > 0 {
		price = item.Product.PromotionPrice
	}

	bundles := globalBundles
	if len(item.Product.Bundles) > 0 {
		bundles = item.Product.Bundles
	}

	// Find the best discount applicable for the quantity
	var discount int
	for _, b := range bundles {
		if item.Quantity >= b.Quantity && b.DiscountPercentage > discount {
			discount = b.DiscountPercentage
		}
	}

	total := price.Multiply(item.Quantity)
	if discount > 0 {
		// Calculate discount: total * discount / 100
		reduction := (int64(total) * int64(discount)) / 100
		total = total - models.Currency(reduction)
	}

	return total
}

// CalculateCartTotal calculates the total price of all items in the cart.
func CalculateCartTotal(cart *models.Cart, globalBundles []models.Bundle) models.Currency {
	var total models.Currency
	for _, item := range cart.Items {
		total += CalculateItemPrice(item, globalBundles)
	}
	return total
}

// CalculateBundlePrice calculates the total price for a product when bought in a specific bundle quantity.
func CalculateBundlePrice(p models.Product, b models.Bundle) models.Currency {
	price := p.Price
	if p.PromotionPrice > 0 {
		price = p.PromotionPrice
	}
	total := price.Multiply(b.Quantity)
	reduction := (int64(total) * int64(b.DiscountPercentage)) / 100
	return total - models.Currency(reduction)
}
