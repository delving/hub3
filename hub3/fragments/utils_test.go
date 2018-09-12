// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fragments

//func TestObjectTypeConversions(t *testing.T) {
//gt := g.NewGomegaWithT(t)
//for s, i := range ObjectXSDType_value {
//var ok bool
//origInt, ok := ObjectXSDType_value[s]
//gt.Expect(ok).To(g.BeTrue())
//gt.Expect(origInt).To(g.Equal(i))

//ot, ok := int2ObjectXSDType[i]
//gt.Expect(ok).To(g.BeTrue())
//gt.Expect(ot).ToNot(g.BeNil())

//label, err := ot.GetLabel()
//gt.Expect(err).ToNot(g.HaveOccurred())
//gt.Expect(label).ToNot(g.BeEmpty())
//gt.Expect(label).To(g.HavePrefix("%s", "http://www.w3.org/2001/XMLSchema#"))

//rawLabel, ok := objectXSDType2XSDLabel[i]
//gt.Expect(ok).To(g.BeTrue())
//gt.Expect(rawLabel).ToNot(g.BeNil())
//gt.Expect(rawLabel).To(g.Equal(label))

//ot2, err := GetObjectXSDType(label)
//gt.Expect(err).ToNot(g.HaveOccurred())
//gt.Expect(ot2).ToNot(g.BeNil())
//gt.Expect(ot).To(g.Equal(ot2))

//s2, ok := ObjectXSDType_name[int32(ot2)]
//gt.Expect(ok).To(g.BeTrue())
//gt.Expect(s2).ToNot(g.BeNil())
//gt.Expect(s).To(g.Equal(s))
//}
//}
