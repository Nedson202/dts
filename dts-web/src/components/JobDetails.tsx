import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Heading, Text, Button, Flex, Card, TextField, Dialog, Grid, Box } from '@radix-ui/themes';
import { getJobDetails, deleteJob, scheduleJob, cancelScheduledJob, updateJob } from '../services/jobService';
import { JobEditDialog } from './JobEditDialog';
import ScheduledJobsList from './ScheduledJobsList';
import { Job } from '../types';
import { format } from 'date-fns';

const JobDetails: React.FC = () => {
    const [job, setJob] = useState<Job | null>(null);
    const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
    const [showScheduleDialog, setShowScheduleDialog] = useState(false);
    const [cpu, setCpu] = useState(1);
    const [memory, setMemory] = useState(1);
    const [storage, setStorage] = useState(1);

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

    const handleSaveJob = async (updatedJob: Job) => {
        if (id) {
            try {
                await updateJob(id, updatedJob);
                setJob(updatedJob);
                setIsEditDialogOpen(false);
                fetchJobDetails(); // Refresh job details after update
            } catch (error) {
                console.error("Failed to update job:", error);
                // Handle error (e.g., show error message to user)
            }
        }
    };

    const handleScheduleJob = async () => {
        if (id) {
            try {
                await scheduleJob(id, cpu, memory, storage);
                setShowScheduleDialog(false);
                fetchJobDetails(); // Refresh job details after scheduling
            } catch (error) {
                console.error("Failed to schedule job:", error);
                // Handle error (e.g., show error message to user)
            }
        }
    };

    const handleCancelScheduledJob = async () => {
        if (id) {
            try {
                await cancelScheduledJob(id);
                fetchJobDetails(); // Refresh job details after cancellation
            } catch (error) {
                console.error("Failed to cancel scheduled job:", error);
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
                        <Text>{job.lastRun ? format(new Date(job.lastRun), 'yyyy-MM-dd HH:mm:ss') : 'N/A'}</Text>
                        <Text weight="bold">Next Run:</Text>
                        <Text>{job.nextRun ? format(new Date(job.nextRun), 'yyyy-MM-dd HH:mm:ss') : 'N/A'}</Text>
                    </Grid>
                </Card>
            </Grid>
            <Box my="6">
                <Flex justify="between" align="center" mb="4">
                    <Heading size="4">Scheduled Runs</Heading>
                    <Button onClick={() => setShowScheduleDialog(true)}>Schedule New Run</Button>
                </Flex>
                <ScheduledJobsList job={job} />
            </Box>

            <JobEditDialog
                isOpen={isEditDialogOpen}
                onClose={() => setIsEditDialogOpen(false)}
                job={job}
                onSave={handleSaveJob}
            />

            <Dialog.Root open={showScheduleDialog} onOpenChange={setShowScheduleDialog}>
                <Dialog.Content style={{ maxWidth: 450 }}>
                    <Dialog.Title>Schedule Job</Dialog.Title>
                    <Dialog.Description size="2" mb="4">
                        Set resource requirements for the job.
                    </Dialog.Description>

                    <Flex direction="column" gap="3">
                        <label>
                            <Text as="div" size="2" mb="1" weight="bold">
                                CPU
                            </Text>
                            <TextField.Input
                                value={cpu}
                                onChange={(e) => setCpu(Number(e.target.value))}
                                type="number"
                            />
                        </label>
                        <label>
                            <Text as="div" size="2" mb="1" weight="bold">
                                Memory
                            </Text>
                            <TextField.Input
                                value={memory}
                                onChange={(e) => setMemory(Number(e.target.value))}
                                type="number"
                            />
                        </label>
                        <label>
                            <Text as="div" size="2" mb="1" weight="bold">
                                Storage
                            </Text>
                            <TextField.Input
                                value={storage}
                                onChange={(e) => setStorage(Number(e.target.value))}
                                type="number"
                            />
                        </label>
                    </Flex>

                    <Flex gap="3" mt="4" justify="end">
                        <Dialog.Close>
                            <Button variant="soft" color="gray">
                                Cancel
                            </Button>
                        </Dialog.Close>
                        <Dialog.Close>
                            <Button onClick={handleScheduleJob}>Schedule</Button>
                        </Dialog.Close>
                    </Flex>
                </Dialog.Content>
            </Dialog.Root>
        </Box>
    );
};

export default JobDetails;
