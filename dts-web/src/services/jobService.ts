import axios from 'axios';
import { Job, ScheduledJob } from '../types';

const API_URL = 'http://localhost:8080/v1';
const SCHEDULER_API_URL = 'http://localhost:8081/v1';

interface CreateJobData {
    name: string;
    description: string;
    cron_expression: string;
    priority: number;
    max_retries: number;
    timeout: number;
    metadata?: { [key: string]: string };
}

export const getJobs = async (): Promise<Job[]> => {
    const response = await axios.get(`${API_URL}/jobs`);
    return response.data.jobs;
};

export const createJob = async (jobData: CreateJobData): Promise<{ jobId: string }> => {
    const response = await axios.post(`${API_URL}/jobs`, jobData);
    return response.data;
};

export const getJobDetails = async (jobId: string): Promise<Job> => {
    const response = await axios.get(`${API_URL}/jobs/${jobId}`);
    return response.data;
};

export const updateJob = async (jobId: string, jobData: Partial<CreateJobData>): Promise<Job> => {
    const response = await axios.put(`${API_URL}/jobs/${jobId}`, jobData);
    return response.data;
};

export const deleteJob = async (jobId: string): Promise<void> => {
    await axios.delete(`${API_URL}/jobs/${jobId}`);
};

export const getScheduledJobs = async (): Promise<ScheduledJob[]> => {
    const response = await axios.get(`${SCHEDULER_API_URL}/scheduler/jobs`);
    return response.data.jobs;
};

export const scheduleJob = async (jobId: string, cpu: number, memory: number, storage: number): Promise<ScheduledJob> => {
    const response = await axios.post(`${SCHEDULER_API_URL}/scheduler/jobs`, { 
        job_id: jobId, 
        resources: { cpu, memory, storage } 
    });
    return response.data;
};

export const cancelScheduledJob = async (scheduledJobId: string): Promise<void> => {
    await axios.delete(`${SCHEDULER_API_URL}/scheduler/jobs/${scheduledJobId}`);
};

export const getScheduledJobDetails = async (scheduledJobId: string): Promise<ScheduledJob> => {
    const response = await axios.get(`${SCHEDULER_API_URL}/scheduler/jobs/${scheduledJobId}`);
    return response.data;
};
