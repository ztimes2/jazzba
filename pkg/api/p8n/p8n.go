package p8n

// Page represents a pagination unit defined by its limit and offset.
type Page struct {
	Limit  int
	Offset int
}

// NewPage initializes a new instance of Page using given limit and offset
// parameters.
func NewPage(limit, offset int) Page {
	sanitizedLimit := limit
	if limit < 0 {
		sanitizedLimit = 0
	}
	sanitizedOffset := offset
	if offset < 0 {
		sanitizedOffset = 0
	}
	return Page{
		Limit:  sanitizedLimit,
		Offset: sanitizedOffset,
	}
}

// Next checks if current page has a next page based on a number of results the
// page contains. A non-nil pointer to a next page is returned in case if the
// page has a next page.
func Next(resultCount int, currentPage Page) *Page {
	if resultCount < currentPage.Limit {
		return nil
	}
	return &Page{
		Limit:  currentPage.Limit,
		Offset: currentPage.Offset + currentPage.Limit,
	}
}

// Previous checks if current page has a previous page. A non-nil pointer to a
// previous page is returned in case if the page has a previous page.
func Previous(currentPage Page) *Page {
	if currentPage.Offset == 0 {
		return nil
	}

	previousLimit := currentPage.Limit
	previousOffset := currentPage.Offset - currentPage.Limit
	if previousOffset < 0 {
		previousLimit = currentPage.Offset
		previousOffset = 0
	}
	return &Page{
		Limit:  previousLimit,
		Offset: previousOffset,
	}
}

// Pagination holds pointers to previous and next pages.
type Pagination struct {
	Next     *Page
	Previous *Page
}

// NewPagination initializes a new Pagination using current page's number of
// results, limit and offset.
func NewPagination(resultCount int, currentPage Page) Pagination {
	return Pagination{
		Next:     Next(resultCount, currentPage),
		Previous: Previous(currentPage),
	}
}

// HasNextPage checks if a pointer to a next page is available.
func (p Pagination) HasNextPage() bool {
	return p.Next != nil
}

// HasPreviousPage checks if a pointer to a previous page is available.
func (p Pagination) HasPreviousPage() bool {
	return p.Previous != nil
}
