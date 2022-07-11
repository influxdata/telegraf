package t128_metrics

// RequestMetric describes one element that will need to be retrieved from the back end
type RequestMetric struct {
	ID             string
	Parameters     []RequestParameter
	OutMeasurement string
	OutField       string
}

// RequestParameter is the simple form of a metric's parameters
type RequestParameter = struct {
	Name    string   `json:"name"`
	Values  []string `json:"values,omitempty"`
	Itemize bool     `json:"itemize"`
}

// ResponseMetric is a single item in this JSON list
// [
//   {
//     "id": "/stats/active-sources",
//     "permutations": [
//       {
//         "parameters": [
//           {
//             "name": "node",
//             "value": "test1"
//           }
//         ],
//         "value": "0"
//       }
//     ]
//   }
// ]
type ResponseMetric struct {
	ID           string                `json:"id"`
	Permutations []ResponsePermutation `json:"permutations"`
}

// ResponsePermutation is a uniquely tagged value within a ResponseMetric
type ResponsePermutation struct {
	Parameters []ResponseParameter `json:"parameters"`
	Value      *string             `json:"value"`
}

// ResponseParameter describes the format of a parameter produced by the 128T REST API
type ResponseParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
