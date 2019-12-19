/*
 * Promethium Daemon API
 *
 * API for Promethium Daemon
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package api

type UpdateVm struct {

	Name string `json:"name,omitempty"`

	Cpus int64 `json:"cpus,omitempty"`

	Memory int64 `json:"memory,omitempty"`
}