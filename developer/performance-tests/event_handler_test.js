/**
 * This script simulates the creation of new pods in the cluster and measures
 * the number of events processed by YHS.
 * It calculates and logs the total number of pods created
 * and the total number of events processed by YHS during the test.
 */
import { Kubernetes } from 'k6/x/kubernetes';
import { sleep } from 'k6';
import http from 'k6/http';

const namespace = __ENV.NAMESPACE || 'default';
const yhsUrl = __ENV.YHS_URL;

const podTemplate = {
    apiVersion: 'v1',
    kind: 'Pod',
    metadata: {
        generateName: 'test-pod-',
        labels: {
            app: 'performance-tests-by-k6'
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
        { duration: '5s', target: 5 },
        { duration: '5s', target: 5 },
        { duration: '5s', target: 0 },
    ],
};

const getStat = () => {
    const res = http.get(`${yhsUrl}/ws/v1/event-statistics`);
    return JSON.parse(res.body);
};

const newEventProcessedByYHS = (data) => getStat()['APP.ADD'] - data.initial_event_count;

export function setup() {
    const kubernetes = new Kubernetes();
    const pods = kubernetes.list("Pod", namespace);
    const stat = getStat();
    const initial_event_count = stat['APP.ADD'] || 0;
    return {
        "initial_pod_count": pods.length,
        "initial_event_count": initial_event_count
    };
}

export default function () {
    const kubernetes = new Kubernetes();
    try {
        kubernetes.create(podTemplate);
    } catch (error) {
        console.log(`Failed to create pod: ${error}`);
    }
    sleep(0.1);
}

export function teardown(data) {
    const kubernetes = new Kubernetes();

    for (let i = 0; i < 5; i++) {
        const pods = kubernetes.list("Pod", namespace);
        const totalPodsSubmitted = pods.length - data.initial_pod_count;
        sleep(i*1);
        console.log(`\nSummary of Load Test at ${i*1}th seconds`);
        console.log(`Number of NEW pods submitted: ${totalPodsSubmitted}`);
        console.log(`Number of NEW Event processed by YHS: ${newEventProcessedByYHS(data)}`);
        console.log(`====================================================\n`);
    }
}