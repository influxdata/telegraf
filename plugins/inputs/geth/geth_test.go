package geth

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMetrics(t *testing.T) {
	var acc testutil.Accumulator
	g := &Geth{
		Metrics: []string{
			"chain",
			"db",
			"discv5",
			"eth",
			"les",
			"p2p",
			"system",
			"trie",
			"txpool",
		},
	}

	gather := func(acc telegraf.Accumulator) error {
		g.parseJSONMetrics(acc, []byte(exampleMetrics), nil)
		return nil
	}

	require.NoError(t, acc.GatherError(gather))

	intMetrics := []string{
		"chain_inserts_meanrate",
		"eth_db_chaindata_compact_time_overall",
		"p2p_outboundconnects_avgrate05min",
		"system_memory_allocs_avgrate15min",
		"trie_memcache_flush_size_avgrate05min",
		"txpool_pending_ratelimit_overall",
	}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasFloatField("geth", metric), metric)
	}
}

const exampleMetrics = `
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "chain": {
      "inserts": {
        "AvgRate01Min": 0,
        "AvgRate05Min": 0,
        "AvgRate15Min": 0,
        "MeanRate": 0,
        "Overall": 0,
        "Percentiles": {
          "20": 0,
          "5": 0,
          "50": 0,
          "80": 0,
          "95": 0
        }
      }
    },
    "db": {
      "preimage": {
        "hits": {
          "Overall": 0
        },
        "total": {
          "Overall": 0
        }
      }
    },
    "discv5": {
      "InboundTraffic": {
        "AvgRate01Min": 0,
        "AvgRate05Min": 0,
        "AvgRate15Min": 0,
        "MeanRate": 0,
        "Overall": 0
      },
      "OutboundTraffic": {
        "AvgRate01Min": 0,
        "AvgRate05Min": 0,
        "AvgRate15Min": 0,
        "MeanRate": 0,
        "Overall": 0
      }
    },
    "eth": {
      "db": {
        "chaindata": {
          "compact": {
            "input": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "output": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "time": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "writedelay": {
              "counter": {
                "AvgRate01Min": 0,
                "AvgRate05Min": 0,
                "AvgRate15Min": 0,
                "MeanRate": 0,
                "Overall": 0
              },
              "duration": {
                "AvgRate01Min": 0,
                "AvgRate05Min": 0,
                "AvgRate15Min": 0,
                "MeanRate": 0,
                "Overall": 0
              }
            }
          },
          "disk": {
            "read": {
              "AvgRate01Min": 274.99971983547925,
              "AvgRate05Min": 438.53373674289816,
              "AvgRate15Min": 474.00341358744504,
              "MeanRate": 58.62976299991441,
              "Overall": 2464
            },
            "write": {
              "AvgRate01Min": 788.8627788368003,
              "AvgRate05Min": 567.257191113575,
              "AvgRate15Min": 519.8551709324448,
              "MeanRate": 1097.3806300776057,
              "Overall": 46119
            }
          }
        }
      },
      "downloader": {
        "bodies": {
          "drop": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "in": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "req": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0,
            "Percentiles": {
              "20": 0,
              "5": 0,
              "50": 0,
              "80": 0,
              "95": 0
            }
          },
          "timeout": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          }
        },
        "headers": {
          "drop": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "in": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "req": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0,
            "Percentiles": {
              "20": 0,
              "5": 0,
              "50": 0,
              "80": 0,
              "95": 0
            }
          },
          "timeout": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          }
        },
        "receipts": {
          "drop": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "in": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "req": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0,
            "Percentiles": {
              "20": 0,
              "5": 0,
              "50": 0,
              "80": 0,
              "95": 0
            }
          },
          "timeout": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          }
        },
        "states": {
          "drop": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "in": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          }
        }
      },
      "fetcher": {
        "fetch": {
          "bodies": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "headers": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          }
        },
        "filter": {
          "bodies": {
            "in": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "out": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          },
          "headers": {
            "in": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "out": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          }
        },
        "prop": {
          "announces": {
            "dos": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "drop": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "in": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "out": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0,
              "Percentiles": {
                "20": 0,
                "5": 0,
                "50": 0,
                "80": 0,
                "95": 0
              }
            }
          },
          "broadcasts": {
            "dos": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "drop": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "in": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "out": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0,
              "Percentiles": {
                "20": 0,
                "5": 0,
                "50": 0,
                "80": 0,
                "95": 0
              }
            }
          }
        }
      },
      "misc": {
        "in": {
          "packets": {
            "AvgRate01Min": 0.10793835279391097,
            "AvgRate05Min": 0.025532896070349136,
            "AvgRate15Min": 0.008760495934007656,
            "MeanRate": 0.20006356619688775,
            "Overall": 8
          },
          "traffic": {
            "AvgRate01Min": 8.512032287751767,
            "AvgRate05Min": 2.013759583611237,
            "AvgRate15Min": 0.6909658191868576,
            "MeanRate": 15.77966423205203,
            "Overall": 631
          }
        },
        "out": {
          "packets": {
            "AvgRate01Min": 0.5017783040534761,
            "AvgRate05Min": 0.120699016767223,
            "AvgRate15Min": 0.041544530550461256,
            "MeanRate": 0.9502834743118045,
            "Overall": 38
          },
          "traffic": {
            "AvgRate01Min": 35.6262595877968,
            "AvgRate05Min": 8.569630190472832,
            "AvgRate15Min": 2.9496616690827486,
            "MeanRate": 67.47011857730797,
            "Overall": 2698
          }
        }
      },
      "prop": {
        "blocks": {
          "in": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          },
          "out": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          }
        },
        "hashes": {
          "in": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          },
          "out": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          }
        },
        "txns": {
          "in": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          },
          "out": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          }
        }
      },
      "req": {
        "bodies": {
          "in": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          },
          "out": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          }
        },
        "headers": {
          "in": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          },
          "out": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          }
        },
        "receipts": {
          "in": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          },
          "out": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          }
        },
        "states": {
          "in": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          },
          "out": {
            "packets": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            },
            "traffic": {
              "AvgRate01Min": 0,
              "AvgRate05Min": 0,
              "AvgRate15Min": 0,
              "MeanRate": 0,
              "Overall": 0
            }
          }
        }
      }
    },
    "les": {
      "misc": {
        "in": {
          "packets": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "traffic": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          }
        },
        "out": {
          "packets": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "traffic": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          }
        }
      }
    },
    "p2p": {
      "InboundConnects": {
        "AvgRate01Min": 0,
        "AvgRate05Min": 0,
        "AvgRate15Min": 0,
        "MeanRate": 0,
        "Overall": 0
      },
      "InboundTraffic": {
        "AvgRate01Min": 569.7115782785817,
        "AvgRate05Min": 136.63011743920686,
        "AvgRate15Min": 47.00898579625933,
        "MeanRate": 1075.0197063190199,
        "Overall": 42990
      },
      "OutboundConnects": {
        "AvgRate01Min": 0.9591242422115045,
        "AvgRate05Min": 0.23159985079537784,
        "AvgRate15Min": 0.07977761793646979,
        "MeanRate": 1.8254785263772082,
        "Overall": 73
      },
      "OutboundTraffic": {
        "AvgRate01Min": 745.8129172308428,
        "AvgRate05Min": 179.5047751197371,
        "AvgRate15Min": 61.79741197402309,
        "MeanRate": 1413.6349326583788,
        "Overall": 56531
      }
    },
    "system": {
      "disk": {
        "readcount": {
          "AvgRate01Min": 11.106675772084918,
          "AvgRate05Min": 5.653028758527626,
          "AvgRate15Min": 4.565742985836539,
          "MeanRate": 16.20234982893697,
          "Overall": 681
        },
        "readdata": {
          "AvgRate01Min": 1016.257528425042,
          "AvgRate05Min": 797.2440193010779,
          "AvgRate15Min": 756.2270787898954,
          "MeanRate": 1157.7898031291963,
          "Overall": 48663
        },
        "writecount": {
          "AvgRate01Min": 46.764932956487215,
          "AvgRate05Min": 24.947562464662326,
          "AvgRate15Min": 20.149849208626904,
          "MeanRate": 78.44229877885518,
          "Overall": 3297
        },
        "writedata": {
          "AvgRate01Min": 5050.742028103304,
          "AvgRate05Min": 2699.021674585294,
          "AvgRate15Min": 2177.8365652506513,
          "MeanRate": 8578.40337902081,
          "Overall": 360558
        }
      },
      "memory": {
        "allocs": {
          "AvgRate01Min": 6140.43030315125,
          "AvgRate05Min": 5976.790114931037,
          "AvgRate15Min": 5926.856283260245,
          "MeanRate": 6656.10492611078,
          "Overall": 279762
        },
        "frees": {
          "AvgRate01Min": 1820.8595831447703,
          "AvgRate05Min": 2435.583728863325,
          "AvgRate15Min": 2557.275561516998,
          "MeanRate": 1143.6827642803373,
          "Overall": 48070
        },
        "inuse": {
          "AvgRate01Min": 43051789.046715096,
          "AvgRate05Min": 71579234.01142918,
          "AvgRate15Min": 77891500.69185199,
          "MeanRate": 3913432.1788982605,
          "Overall": 164485048
        },
        "pauses": {
          "AvgRate01Min": 515472.3270178064,
          "AvgRate05Min": 817308.5211633872,
          "AvgRate15Min": 882600.2430201655,
          "MeanRate": 118070.28501421161,
          "Overall": 4962600
        }
      }
    },
    "trie": {
      "cachemiss": {
        "Overall": 4
      },
      "cacheunload": {
        "Overall": 0
      },
      "memcache": {
        "commit": {
          "nodes": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "size": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "time": {
            "Mean": "0s",
            "Measurements": 0,
            "Percentiles": {
              "20": "0s",
              "5": "0s",
              "50": "0s",
              "80": "0s",
              "95": "0s"
            }
          }
        },
        "flush": {
          "nodes": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "size": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "time": {
            "Mean": "0s",
            "Measurements": 0,
            "Percentiles": {
              "20": "0s",
              "5": "0s",
              "50": "0s",
              "80": "0s",
              "95": "0s"
            }
          }
        },
        "gc": {
          "nodes": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "size": {
            "AvgRate01Min": 0,
            "AvgRate05Min": 0,
            "AvgRate15Min": 0,
            "MeanRate": 0,
            "Overall": 0
          },
          "time": {
            "Mean": "0s",
            "Measurements": 0,
            "Percentiles": {
              "20": "0s",
              "5": "0s",
              "50": "0s",
              "80": "0s",
              "95": "0s"
            }
          }
        }
      }
    },
    "txpool": {
      "invalid": {
        "Overall": 0
      },
      "pending": {
        "discard": {
          "Overall": 0
        },
        "nofunds": {
          "Overall": 0
        },
        "ratelimit": {
          "Overall": 0
        },
        "replace": {
          "Overall": 0
        }
      },
      "queued": {
        "discard": {
          "Overall": 0
        },
        "nofunds": {
          "Overall": 0
        },
        "ratelimit": {
          "Overall": 0
        },
        "replace": {
          "Overall": 0
        }
      },
      "underpriced": {
        "Overall": 0
      }
    }
  }
}
`
