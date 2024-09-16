import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Table, Button, Heading } from '@radix-ui/themes';
import { getJobs } from '../services/jobService';

interface Job {
    id: string;
    name: string;
    description: string;
    status: string;
    createdAt: string;
    updatedAt: string;
    cronExpression: string;
    metadata: { [key: string]: string };
    priority: number;
    max_retries: number;
    timeout: number;
}

const JobList: React.FC = () => {
    const [jobs, setJobs] = useState<Job[]>([]);

    useEffect(() => {
        const fetchJobs = async () => {
            const fetchedJobs = await getJobs();
            setJobs(fetchedJobs);
        };
        fetchJobs();
    }, []);

    return (
        <div style={{ padding: '20px' }}>
            <Heading size="4" style={{ marginBottom: '20px' }}>Job List</Heading>
            <Button asChild style={{ marginBottom: '20px' }}>
                <Link to="/create">Create New Job</Link>
            </Button>
            <Table.Root>
                <Table.Header>
                    <Table.Row>
                        <Table.ColumnHeaderCell>Job Name</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Status</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Priority</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Created At</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Actions</Table.ColumnHeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {jobs.map(job => (
                        <Table.Row key={job.id}>
                            <Table.Cell>{job.name}</Table.Cell>
                            <Table.Cell>{job.status}</Table.Cell>
                            <Table.Cell>{job.priority}</Table.Cell>
                            <Table.Cell>{new Date(job.createdAt).toLocaleString()}</Table.Cell>
                            <Table.Cell>
                                <Button asChild variant="soft">
                                    <Link to={`/job/${job.id}`}>View Details</Link>
                                </Button>
                            </Table.Cell>
                        </Table.Row>
                    ))}
                </Table.Body>
            </Table.Root>
        </div>
    );
};

export default JobList;
