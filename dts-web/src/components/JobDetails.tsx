import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Heading, Text, Button, Flex, Card, Grid, Box } from '@radix-ui/themes';
import { getJobDetails, deleteJob, updateJob, cancelJob } from '../services/jobService';
import { JobEditDialog } from './JobEditDialog';
import ExecutionHistoryList from './ExecutionHistoryList';
import { Job, JobEdit } from '../types';
import { format } from 'date-fns';

const JobDetails: React.FC = () => {
    const [job, setJob] = useState<Job | null>(null);
    const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);

    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();

    useEffect(() => {
        fetchJobDetails();
    }, [id]);

    const fetchJobDetails = async () => {
        if (id) {
            try {
                const jobDetails = await getJobDetails(id);
                setJob(jobDetails);
            } catch (error) {
                console.error("Failed to fetch job details:", error);
                // Handle error (e.g., show error message to user)
            }
        }
    };

    const handleDelete = async () => {
        if (id) {
            try {
                await deleteJob(id);
                navigate('/');
            } catch (error) {
                console.error("Failed to delete job:", error);
                // Handle error (e.g., show error message to user)
            }
        }
    };

    const handleEditJob = () => {
        setIsEditDialogOpen(true);
    };

    const handleSaveEditJob = async (updatedJob: JobEdit) => {
        if (id && job) {
            try {
                await updateJob(id, updatedJob);
                setIsEditDialogOpen(false);
                fetchJobDetails(); // Refresh job details after update
            } catch (error) {
                console.error("Failed to update job:", error);
                // Handle error (e.g., show error message to user)
            }
        }
    };

    const handleCancelJob = async () => {
        if (job && job.id) {
            try {
                await cancelJob(job.id);
                fetchJobDetails(); // Refresh job details after cancellation
            } catch (error) {
                console.error("Failed to cancel job:", error);
                // Handle error (e.g., show error message to user)
            }
        }
    };

    if (!job) {
        return <Text>Loading...</Text>;
    }

    return (
        <Box p="4">
            <Flex justify="between" align="center" mb="4">
                <Heading size="6">{job.name}</Heading>
                <Flex gap="2">
                    <Button onClick={handleEditJob}>Edit</Button>
                    <Button color="red" onClick={handleDelete}>Delete</Button>
                    <Button color="yellow" onClick={handleCancelJob} disabled={job.status === 'COMPLETED' || job.status === 'FAILED' || job.status === 'CANCELLED'}>
                        Cancel
                    </Button>
                </Flex>
            </Flex>
            <Grid columns="2" gap="4">
                <Card>
                    <Heading size="3" mb="2">Job Details</Heading>
                    <Grid columns="2" gap="2">
                        <Text weight="bold">Description:</Text>
                        <Text>{job.description}</Text>
                        <Text weight="bold">Status:</Text>
                        <Text>{job.status}</Text>
                        <Text weight="bold">Cron Expression:</Text>
                        <Text>{job.cronExpression}</Text>
                        <Text weight="bold">Priority:</Text>
                        <Text>{job.priority}</Text>
                        <Text weight="bold">Max Retries:</Text>
                        <Text>{job.maxRetries}</Text>
                        <Text weight="bold">Timeout:</Text>
                        <Text>{job.timeout} seconds</Text>
                    </Grid>
                </Card>
                <Card>
                    <Heading size="3" mb="2">Execution Info</Heading>
                    <Grid columns="2" gap="2">
                        <Text weight="bold">Created At:</Text>
                        <Text>{format(new Date(job.createdAt), 'yyyy-MM-dd HH:mm:ss')}</Text>
                        <Text weight="bold">Updated At:</Text>
                        <Text>{format(new Date(job.updatedAt), 'yyyy-MM-dd HH:mm:ss')}</Text>
                        <Text weight="bold">Last Run:</Text>
                        <Text>{job.lastRun ? format(new Date(job.lastRun), 'yyyy-MM-dd HH:mm:ss') : 'Never'}</Text>
                        <Text weight="bold">Next Run:</Text>
                        <Text>{job.nextRun ? format(new Date(job.nextRun), 'yyyy-MM-dd HH:mm:ss') : 'N/A'}</Text>
                    </Grid>
                </Card>
            </Grid>
            <Box my="6">
                <ExecutionHistoryList job={job} />
            </Box>

            <JobEditDialog
                isOpen={isEditDialogOpen}
                onClose={() => setIsEditDialogOpen(false)}
                job={job}
                onSave={handleSaveEditJob}
            />
        </Box>
    );
};

export default JobDetails;
