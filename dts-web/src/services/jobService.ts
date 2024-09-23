import axios from 'axios';
import { Job } from '../types';

const API_URL = 'http://localhost:8080/v1';

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

