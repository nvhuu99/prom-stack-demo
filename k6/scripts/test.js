import http from "k6/http"

export const options = {
  scenarios: {
    demo_web: {
      executor: "constant-arrival-rate",
      rate: 10,
      timeUnit: "1s",
      duration: "1h",
      preAllocatedVUs: 10,
      maxVUs: 50,
    },
  },
  // thresholds: {
  //   http_req_failed: [
  //     { threshold: "rate<0.1", abortOnFail: true },
  //   ],
  // },
}

export default function () {
  http.get(`http://demo-service:8080/`)
}
