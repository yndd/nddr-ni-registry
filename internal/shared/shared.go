/*
Copyright 2021 NDD.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package shared

import (
	"time"

	"github.com/yndd/ndd-runtime/pkg/logging"
	"github.com/yndd/nddr-ni-registry/internal/handler"
	"github.com/yndd/nddr-organization/pkg/registry"
)

type NddControllerOptions struct {
	Logger    logging.Logger
	Poll      time.Duration
	Namespace string
	Handler   handler.Handler
	Registry  registry.Registry
}
