// 运行示例：PRJ-17
package main

import "fmt"

func action(status, existing string) string {
	switch status {
	case "Success":
		return "accepted"
	case "Publish Timeout":
		return "wait_visibility"
	case "Label Already Exists":
		if existing == "FINISHED" {
			return "verify_original_batch"
		}
		return "query_existing_job"
	default:
		return "inspect_failure"
	}
}

func main() {
	fmt.Println(action("Success", ""))
	fmt.Println(action("Publish Timeout", ""))
	fmt.Println(action("Label Already Exists", "FINISHED"))
	fmt.Println(action("Fail", ""))
}
