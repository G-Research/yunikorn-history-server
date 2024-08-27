export interface ClusterInfo {
    startTime: number;
    rmBuildInformation: BuildInfo[];
    partition: string;
    clusterName: string;
    clusterStatus?: string;
}

export interface BuildInfo {
    buildDate: string;
    buildVersion: string;
    isPluginVersion: string,
    rmId: string;
}
