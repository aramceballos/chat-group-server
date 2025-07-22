import http from "k6/http";
import { sleep, check } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 200 },
    { duration: "5m", target: 200 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(99)<100']
  }
};

export default () => {
  let res = http.get("http://localhost:4000/api/v1/users");
  check(res, "200", (r) => r.status === 200);
  sleep(1);
};
