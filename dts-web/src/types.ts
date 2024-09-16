
export interface Job {
    id: string;
    name: string;
    description: string;
    status: string;
    createdAt: string;
    updatedAt: string;
    cronExpression: string;
    metadata: { [key: string]: string };
    priority: number;
    maxRetries: number;
    timeout: number;
}
