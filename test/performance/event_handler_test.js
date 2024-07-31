import { sleep } from 'k6';
import http from 'k6/http';
import { Counter } from 'k6/metrics';
import { Kubernetes } from 'k6/x/kubernetes';

const jobCounter = new Counter('job_counter');
const jobErrorCounter = new Counter('job_error_counter');
const appAddCounter = new Counter('app_add_counter');

const namespace = __ENV.NAMESPACE || 'default';
const yhsServer = __ENV.YHS_SERVER || 'http://localhost:8989';

const EVENT_STATISTICS_PATH = '/ws/v1/event-statistics';
const jobTemplate = {
    apiVersion: 'batch/v1',
    kind: 'Job',
    metadata: {
        generateName: 'batch-sleep-job-',
        namespace: namespace,
    },
    spec: {
        ttlSecondsAfterFinished: 15,
        template: {
            metadata: {
                labels: {
                    app: 'sleep',
                    applicationId: 'batch-sleep-job-1',
                    queue: 'root.sandbox',
                },
            },
            spec: {
                schedulerName: 'yunikorn',
                restartPolicy: 'Never',
                containers: [{
                    name: 'sleep',
                    image: 'alpine:latest',
                    command: ['sleep', '10'],
                }],
            },
        },
    },
};

// Function to create a Job in Kubernetes
function createJob(kubernetes) {
    try {
        kubernetes.create(jobTemplate);
        jobCounter.add(1);
    } catch (error) {
        jobErrorCounter.add(1);
        console.error(`Failed to create job: ${error}`);
    }
}

// Function to get statistics from YHS server
function getStats() {
    const res = http.get(`${yhsServer}${EVENT_STATISTICS_PATH}`);
    if (res.status !== 200) {
        console.error(`Failed to fetch statistics: ${res.status} ${res.statusText}`);
        return null;
    }
    return JSON.parse(res.body);
}

// Test setup
export const options = {
    stages: [
        { duration: '120s', target: 100 },
        { duration: '360s', target: 350 },
        { duration: '120s', target: 0 },
    ],
};

// Test execution
export default function () {
    const kubernetes = new Kubernetes();
    createJob(kubernetes);
    sleep(5)
    const stat = getStats();
    appAddCounter.add(stat["APP-ADD"])
    sleep(5)
}
