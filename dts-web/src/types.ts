export interface Job {
    id: string;
    name: string;
    description: string;
    status: string;
    createdAt: string;
    updatedAt: string;
    cronExpression: string;
    priority: number;
    maxRetries: number;
    timeout: number;
    metadata: Record<string, string>;
    nextRun: string;
    lastRun: string;
}

export interface Resources {
    cpu: number;
    memory: number;
    storage: number;
}

export interface ScheduledJob {
    jobId: string;
    job: Job;
    resourceRequirements: Resources;
    nextExecutionTime: string;
}
