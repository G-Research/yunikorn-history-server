import { Kubernetes } from 'k6/x/kubernetes';
import http from 'k6/http';
import { Trend } from 'k6/metrics';

let podsSubmitted = new Trend('pods_submitted');
let eventsProcessed = new Trend('events_processed');

const namespace = __ENV.NAMESPACE || 'default';
const yhsServer = __ENV.YHS_SERVER || 'http://localhost:8989';

const podTemplate = {
    apiVersion: 'v1',
    kind: 'Pod',
    metadata: {
        generateName: 'test-pod-',
        labels: {
            app: 'performance-by-k6'
        },
        namespace: namespace
    },
    spec: {
        containers: [
            {
                name: 'test-container',
                image: 'nginx:latest',
                ports: [{ containerPort: 80 }]
            }
        ]
    }
};

export const options = {
    stages: [
        { duration: '15s', target: 5 },
        { duration: '15s', target: 5 },
        { duration: '15s', target: 0 },
    ],
};

const getStat = () => {
    const res = http.get(`${yhsServer}/ws/v1/event-statistics`);
    return JSON.parse(res.body);
};

export default function () {
    const kubernetes = new Kubernetes();
    try {
        kubernetes.create(podTemplate);
    } catch (error) {
        console.log(`Failed to create pod: ${error}`);
    }
    const json = getStat();
    const eventProcessedByYHS = json["APP-ADD"]
    const pods = kubernetes.list("Pod", namespace);
    // Add the number of pods submitted and the number of events processed to the metrics
    podsSubmitted.add(pods.length);
    eventsProcessed.add(eventProcessedByYHS);
}
