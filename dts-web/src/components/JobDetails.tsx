import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Heading, Text, Button, Flex, Card, Table } from '@radix-ui/themes';
import { getJobDetails, deleteJob } from '../services/jobService';
import { JobEditDialog } from './JobEditDialog';
import { Job } from '../types';


const JobDetails: React.FC = () => {
    const [job, setJob] = useState<Job | null>(null);
    const [selectedJob, setSelectedJob] = useState<Job | null>(null);

    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();

    useEffect(() => {
        const fetchJobDetails = async () => {
            if (id) {
                const jobDetails = await getJobDetails(id);
                setJob(jobDetails);
            }
        };
        fetchJobDetails();
    }, [id, selectedJob]);

    const handleDelete = async () => {
        if (id) {
            await deleteJob(id);
            navigate('/');
        }
    };

    const handleEditJob = (e: React.FormEvent, job: Job) => {
        e.preventDefault();
        setSelectedJob(job);
    };

    const handleSaveJob = (updatedJob: Job) => {
        // Here you would typically also make an API call to update the job on the server
    };

    if (!job) {
        return <Text>Loading...</Text>;
    }

    return (
        <Flex direction="column" gap="4" style={{ padding: '20px' }}>
            <Heading size="4">{job.name}</Heading>
            <Card>
                <Flex direction="column" gap="2">
                    <Text>Description: {job.description}</Text>
                    <Text>Status: {job.status}</Text>
                    <Text>Created At: {new Date(job.createdAt).toLocaleString()}</Text>
                    <Text>Updated At: {new Date(job.updatedAt).toLocaleString()}</Text>
                    <Text>Cron Expression: {job.cronExpression}</Text>
                    <Text>Priority: {job.priority}</Text>
                    <Text>Max Retries: {job.maxRetries}</Text>
                    <Text>Timeout: {job.timeout} seconds</Text>
                </Flex>
            </Card>
            <Heading size="3">Metadata</Heading>
            <Table.Root>
                <Table.Header>
                    <Table.Row>
                        <Table.ColumnHeaderCell>Key</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Value</Table.ColumnHeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {Object.entries(job.metadata).map(([key, value]) => (
                        <Table.Row key={key}>
                            <Table.Cell>{key}</Table.Cell>
                            <Table.Cell>{value}</Table.Cell>
                        </Table.Row>
                    ))}
                </Table.Body>
            </Table.Root>

            <JobEditDialog
                job={selectedJob}
                onSave={handleSaveJob}
                onClose={() => setSelectedJob(null)}
            />
            <Flex gap="2">
                <Button onClick={(e) => handleEditJob(e, job)}>Edit</Button>
                <Button color="red" onClick={handleDelete}>Delete</Button>
            </Flex>
        </Flex>
    );
};

export default JobDetails;
