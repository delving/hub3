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

// IsPhraseMatch is a helper to determine if two positions are close enough to
// be part of the same phrase. This is used for phrase queries. Default slop is 0.
func IsPhraseMatch(pos1, pos2, slop int) (bool, error) {
	// a valid position can occur before or after pos1 within the slop
	pos := pos1 - pos2
	if slop == 0 && pos == -1 {
		return true, nil
	}

	upper := slop
	lower := -slop - 1

	if pos >= lower && pos <= upper {
		return true, nil
	}

	return false, nil
}

// ValidPhrasePosition returns a list of valid positions from the source position
// to determine if the term is part of a phrase.
func ValidPhrasePosition(vector Vector, slop int) []Vector {
	if slop == 0 {
		return []Vector{
			vector,
			{DocID: vector.DocID, Location: vector.Location + 1},
		}
	}

	valid := []Vector{}

	pos := vector.Location

	for i := pos - slop; i <= pos+slop; i++ {
		if i > 0 {
			valid = append(valid, Vector{DocID: vector.DocID, Location: i})
		}
	}

	return valid
}
