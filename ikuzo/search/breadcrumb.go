// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package search

import "strings"

// BreadCrumb is a building block for displaying search and additional filtering.
// Each next step of filtering is an additional BreadCrumb.
type BreadCrumb struct {
	Href           string `json:"href,omitempty"`
	Display        string `json:"display,omitempty"`
	Field          Label  `json:"field,omitempty"`
	LocalizedField string `json:"localized_field,omitempty"`
	Value          string `json:"value,omitempty"`
	IsLast         bool   `json:"is_last,omitempty"`
}

// BreadCrumbBuilder holds all the information to build a search breadcrumb trail.
type BreadCrumbBuilder struct {
	hrefPath []string
	crumbs   []*BreadCrumb
}

// GetLast returns the last BreadCrumb from the trail
func (bcb *BreadCrumbBuilder) GetLast() *BreadCrumb {
	if len(bcb.crumbs) == 0 {
		return nil
	}

	return bcb.crumbs[len(bcb.crumbs)-1]
}

// GetPath returns the path for the BreadCrumb
func (bcb *BreadCrumbBuilder) GetPath() string {
	return strings.Join(bcb.hrefPath, "&")
}

// AppendBreadCrumb creates a BreadCrumb
// func (bcb *BreadCrumbBuilder) AppendBreadCrumb(param string, qf *QueryFilter) {
// bc := &BreadCrumb{IsLast: true}
// switch param {
// case "query":
// if qf.GetValue() != "" {
// bc.Display = qf.GetValue()
// bc.Href = fmt.Sprintf("q=%s", qf.GetValue())
// bc.Value = qf.GetValue()
// bcb.hrefPath = append(bcb.hrefPath, bc.Href)
// }
// case "qf[]", qfKey, qfIDKey, "qf.id[]":
// if !strings.HasSuffix(param, "[]") {
// param = fmt.Sprintf("%s[]", param)
// }
// qfs := fmt.Sprintf("%s:%s", qf.GetSearchLabel(), qf.GetValue())
// if qf.Exclude {
// qfs = fmt.Sprintf("-%s", qfs)
// }
// href := fmt.Sprintf("%s=%s", param, qfs)
// bc.Href = href
// if bcb.GetPath() != "" {
// bc.Href = bcb.GetPath() + "&" + bc.Href
// }
// bcb.hrefPath = append(bcb.hrefPath, href)
// bc.Display = qfs
// bc.Field = qf.GetSearchLabel()
// bc.Value = qf.GetValue()
// case "qf.exist[]", "qf.exist":
// if !strings.HasSuffix(param, "[]") {
// param = fmt.Sprintf("%s[]", param)
// }
// qfs := qf.GetSearchLabel()
// href := fmt.Sprintf("%s=%s", param, qfs)
// bc.Href = href
// if bcb.GetPath() != "" {
// bc.Href = bcb.GetPath() + "&" + bc.Href
// }
// bcb.hrefPath = append(bcb.hrefPath, href)
// bc.Display = qfs
// bc.Field = qf.GetSearchLabel()
// // bc.Value = qf.GetValue()
// }
// last := bcb.GetLast()
// if last != nil {
// last.IsLast = false
// }
// bcb.crumbs = append(bcb.crumbs, bc)
// }
