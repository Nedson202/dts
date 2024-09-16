import axios from 'axios';

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

interface JobData extends CreateJobData {
    id: string;
    status: string;
    created_at: string;
    updated_at: string;
}

export const getJobs = async () => {
    const response = await axios.get(`${API_URL}/jobs`);
    return response.data.jobs;
};

export const createJob = async (jobData: CreateJobData) => {
    const response = await axios.post(`${API_URL}/jobs`, jobData);
    return response.data;
};

export const getJobDetails = async (jobId: string) => {
    const response = await axios.get(`${API_URL}/jobs/${jobId}`);
    return response.data;
};

export const updateJob = async (jobId: string, jobData: Partial<CreateJobData>) => {
    const response = await axios.put(`${API_URL}/jobs/${jobId}`, jobData);
    return response.data;
};

export const deleteJob = async (jobId: string) => {
    await axios.delete(`${API_URL}/jobs/${jobId}`);
};

// Add more functions as needed for other API interactions
