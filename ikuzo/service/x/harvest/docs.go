package harvest

/*
Algorithm

- harvest all records and record last modified and completeListSize
- next harvest
	- get completeListSize and compare to previous count
	- harvest previous until date get count
- generate diff
	- NoChange: nothing changed when changed == 0 and old result is same count
	- Modified: new records and same count is decremented from previous list size
	- New: new records and previous list size is identifal
	- Deleted: no change, previous list size is smaller
	- DeletedNew: new records are added, previous list decremented ? Maybe ambiguous.
		- Size of previous and current list is identical. Or other small change
	- DeletedModified: modified records, but previous list is decremented more than modified
	- ModifiedNew: more added in diff, than are decremented from the previous list

How to find deleted record identifier in the most efficient way. The expectation is that it is removed from the list.
	- get list of ids for new/modified ids
		- optionally create new list of ids sorted by last modified. This is the internal list of how it should be.

	- sort record ids with last modified
	# if smaller than 100 harvest the whole list
	# create binary search option to split the list
	- take half the list make call with api and check count to internal list
		- take middle item from list and do query with until date. if size is identical split latter
			if size smaller take former half. and split that


# requirements for incremental fast update

- results must be sorted from first modified to last modified
- get result sub-set based on date modified: all until date x, all changed from date x
- each record must include an identifier and a last modified timestamp
- timestamp must include minutes, ideally seconds.
	- ISO 8601 YYYY-MM-DDThh:mm:ssZ
- API result gives back total hits for search result
- API supports pagination through results
- output XML or JSON
- nice to have:
	- ideally support deleted or disabled records
*/
