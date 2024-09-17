export class SchedulerHealthInfo {
    Healthy: boolean = false;
    HealthChecks: null | HealthCheckInfo[] = [];
}

export interface HealthCheckInfo {
    Name: string;
	Succeeded: boolean;
	Description: string;
	DiagnosisMessage: string;
}

