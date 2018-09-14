# Webpagetest Input Plugin

This plugin gathers stats from [Webpagetest](https://www.webpagetest.org/).

### Configuration:

```toml
[[inputs.webpagetest]]
  ## WebPageTest API Key
  ## Get from https://www.webpagetest.org/getkey.php
  api_key = "your-api-key"
  ## URLs to test
  urls = ["https://www.example.com"]
  ## Lookup interval. You *probably* want this to run less frequently than
  ## Telegraf's global interval
  interval = "1h"
```

### Measurements & Fields

- webpagetest
  - [`ttfb`](https://sites.google.com/a/webpagetest.org/docs/using-webpagetest/metrics#TOC-First-Byte) - Time to first byte.
  - [`start_render`](https://sites.google.com/a/webpagetest.org/docs/using-webpagetest/metrics#TOC-Start-Render) - Time until first non-white content is painted to the browser display.
  - [`speed_index`](https://sites.google.com/a/webpagetest.org/docs/using-webpagetest/metrics#TOC-Speed-Index) - Time until the visible parts of the page are displayed
  - [`document_complete`](https://sites.google.com/a/webpagetest.org/docs/using-webpagetest/quick-start-quide#TOC-Document-Complete) - window.load event
  - [`fully_loaded`](https://sites.google.com/a/webpagetest.org/docs/using-webpagetest/quick-start-quide#TOC-Fully-Loaded)` - No requests have been made for 2 seconds after window.load event
  - `bytes_in` - Number of bytes transfered before the `fully_loaded` time
  - `bytes_in_doc` - Number of bytes transfered before the `document_complete` time
  - `requests_full` - Number of requests made before the `fully_loaded` time
  - `requests_doc` - Number of requests made before the `document_complete` time
  - `requests_css` - Number of requests for CSS files
  - `bytes_css` - Number of bytes transfered for CSS files
  - `requests_image` - Number of requests for image files
  - `bytes_image` - Number of bytes transfered for image files
  - `requests_js` - Number of requests for JS files
  - `bytes_js` - Number of bytes transfered for JS files
  - `requests_html` - Number of requests for HTML files
  - `bytes_html` - Number of bytes transfered for HTML files
  - `requests_font` - Number of requests for font files
  - `bytes_font` - Number of bytes transfered for font files
  - `requests_other` - Number of requests for "other" files
  - `bytes_other` - Number of bytes transfered for "other" files


### Tags
- url
- type (firstView or repeatView)
