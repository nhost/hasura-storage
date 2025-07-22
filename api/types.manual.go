package api

import "time"

// GetQ returns the Q field value.
func (g GetFileParams) GetQ() *int {
	return g.Q
}

// GetH returns the H field value.
func (g GetFileParams) GetH() *int {
	return g.H
}

// GetW returns the W field value.
func (g GetFileParams) GetW() *int {
	return g.W
}

// GetB returns the B field value.
func (g GetFileParams) GetB() *float32 {
	return g.B
}

// GetF returns the F field value.
func (g GetFileParams) GetF() *OutputImageFormat {
	return g.F
}

// GetIfMatch returns the IfMatch field value.
func (g GetFileParams) GetIfMatch() *string {
	return g.IfMatch
}

// GetIfNoneMatch returns the IfNoneMatch field value.
func (g GetFileParams) GetIfNoneMatch() *string {
	return g.IfNoneMatch
}

// GetIfModifiedSince returns the IfModifiedSince field value.
func (g GetFileParams) GetIfModifiedSince() *time.Time {
	return g.IfModifiedSince
}

// GetIfUnmodifiedSince returns the IfUnmodifiedSince field value.
func (g GetFileParams) GetIfUnmodifiedSince() *time.Time {
	return g.IfUnmodifiedSince
}

// GetQ returns the Q field value.
func (g GetFileMetadataHeadersParams) GetQ() *int {
	return g.Q
}

// GetH returns the H field value.
func (g GetFileMetadataHeadersParams) GetH() *int {
	return g.H
}

// GetW returns the W field value.
func (g GetFileMetadataHeadersParams) GetW() *int {
	return g.W
}

// GetB returns the B field value.
func (g GetFileMetadataHeadersParams) GetB() *float32 {
	return g.B
}

// GetF returns the F field value.
func (g GetFileMetadataHeadersParams) GetF() *OutputImageFormat {
	return g.F
}

// GetIfMatch returns the IfMatch field value.
func (g GetFileMetadataHeadersParams) GetIfMatch() *string {
	return g.IfMatch
}

// GetIfNoneMatch returns the IfNoneMatch field value.
func (g GetFileMetadataHeadersParams) GetIfNoneMatch() *string {
	return g.IfNoneMatch
}

// GetIfModifiedSince returns the IfModifiedSince field value.
func (g GetFileMetadataHeadersParams) GetIfModifiedSince() *time.Time {
	return g.IfModifiedSince
}

// GetIfUnmodifiedSince returns the IfUnmodifiedSince field value.
func (g GetFileMetadataHeadersParams) GetIfUnmodifiedSince() *time.Time {
	return g.IfUnmodifiedSince
}
