/*
Package sitemap provides a service to generated sitemaps

It requires an implementation of the Store interface to render content.

All sitemap formats limit a single sitemap to 10MB (uncompressed) and 50,000 URLs.

For a single Sitemap index file, the maximum capacity of URLs and storage
  could be calculated as described below:

  in terms of URLs:

	50,000 sitemaps = ( 50,000 * 50,000 ) URLs = 2,500,000,000 URLs

  in terms of size:

	50MB + (50,000 sitemaps * 50MB) = 2,500,050 MB = > 2.3 TB

These limits are enforced by the http.Handlers in this package.
*/
package sitemap
