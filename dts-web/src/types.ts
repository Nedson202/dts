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
    lastRun: string | null;
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

export interface Execution {
    id: string;
    jobId: string;
    status: string;
    startTime: string;
    endTime?: string;
    result?: string;
    error?: string;
}

export type JobEdit = Omit<Job, 'createdAt' | 'updatedAt' | 'lastRun' | 'nextRun' | 'status'>;
