# Webpagetest Input Plugin

This plugin gathers stats from [Webpagetest](https://www.webpagetest.org/).

### Configuration:

```toml
[[inputs.webpagetest]]

  ## WebPageTest API Key
  ## Get from https://www.webpagetest.org/getkey.php
  api_key = "key"

  ## URLs to test
  urls = ["https://in.hotels.com/"]

  ## Lookup interval. You *probably* want this to run less frequently than
  interval = "1h"

  ## Since test results are not generated instantaneously,
  # pollFrequency = 5      # Polling frequency in seconds
  # maxPollTime = 120      # Maximum poll/wait time in seconds

  ## Network connectivity information
  ## Refer https://sites.google.com/a/webpagetest.org/docs/advanced-features/webpagetest-restful-apis#TOC-Specifying-connectivity
  # downloadBandwidth = 5000    # kbps
  # uploadBandwidth = 1000      # kbps
  # roundTripLatency = 28       # ms
  # packetLossRate = 0
```

## Metrics
### Measurements & Fields

- webpagetest
  - [`ttfb`](https://sites.google.com/a/webpagetest.org/docs/using-webpagetest/metrics#TOC-First-Byte) - Time to first byte.
  - [`start_render`](https://sites.google.com/a/webpagetest.org/docs/using-webpagetest/metrics#TOC-Start-Render) - Time until first non-white content is painted to the browser display.
  - [`speed_index`](https://sites.google.com/a/webpagetest.org/docs/using-webpagetest/metrics#TOC-Speed-Index) - Time until the visible parts of the page are displayed
  - [`document_complete`](https://sites.google.com/a/webpagetest.org/docs/using-webpagetest/quick-start-quide#TOC-Document-Complete) - window.load event
  - [`fully_loaded`](https://sites.google.com/a/webpagetest.org/docs/using-webpagetest/quick-start-quide#TOC-Fully-Loaded)` - No requests have been made for 2 seconds after window.load event
  - `bytes_in` - Number of bytes transferred before the `fully_loaded` time
  - `bytes_in_doc` - Number of bytes transferred before the `document_complete` time
  - `requests_full` - Number of requests made before the `fully_loaded` time
  - `requests_doc` - Number of requests made before the `document_complete` time
  - `requests_css` - Number of requests for CSS files
  - `bytes_css` - Number of bytes transferred for CSS files
  - `requests_image` - Number of requests for image files
  - `bytes_image` - Number of bytes transferred for image files
  - `requests_js` - Number of requests for JS files
  - `bytes_js` - Number of bytes transferred for JS files
  - `requests_html` - Number of requests for HTML files
  - `bytes_html` - Number of bytes transferred for HTML files
  - `requests_font` - Number of requests for font files
  - `bytes_font` - Number of bytes transferred for font files
  - `requests_other` - Number of requests for "other" files
  - `bytes_other` - Number of bytes transferred for "other" files


### Tags
- url
- type (firstView or repeatView)

### Example Output
```
webpagetest,host=host1,type=firstView,url=https://in.hotels.com/ bytes_css=50117i,bytes_font=58660i,bytes_html=36921i,bytes_image=325394i,bytes_in=1204740i,bytes_in_doc=940975i,bytes_js=721849i,bytes_other=6369i,document_complete=5515i,fully_loaded=9185i,requests_css=1i,requests_doc=49i,requests_font=1i,requests_full=137i,requests_html=20i,requests_image=38i,requests_js=29i,requests_other=24i,speed_index=3508i,start_render=1300i,ttfb=629i 1571472281000000000
webpagetest,host=host1,type=repeatView,url=https://in.hotels.com/ bytes_css=51292i,bytes_font=0i,bytes_html=33044i,bytes_image=908i,bytes_in=316083i,bytes_in_doc=261332i,bytes_js=223622i,bytes_other=7217i,document_complete=2168i,fully_loaded=4685i,requests_css=1i,requests_doc=27i,requests_font=0i,requests_full=105i,requests_html=18i,requests_image=28i,requests_js=12i,requests_other=23i,speed_index=1865i,start_render=900i,ttfb=432i 1571472281000000000
```