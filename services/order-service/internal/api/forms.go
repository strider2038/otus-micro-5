/*
 * Order service
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package api

import "github.com/gofrs/uuid"

type CreateOrderForm struct {
	UserID uuid.UUID `json:"-"`
	Price  float64   `json:"price"`
}