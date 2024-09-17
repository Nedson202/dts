import React, { useState, useEffect } from 'react';
import { Table, Heading, Card, Flex, Text, Button } from '@radix-ui/themes';
import { Job, ScheduledJob } from '../types';
import { getScheduledJobs, cancelScheduledJob } from '../services/jobService';
import { format } from 'date-fns';

const ScheduledJobsList: React.FC<{ job: Job }> = ({ job }) => {
    const [scheduledJobs, setScheduledJobs] = useState<ScheduledJob[]>([]);

    useEffect(() => {
        fetchScheduledJobs();
    }, [job]);

    const fetchScheduledJobs = async () => {
        const jobs = await getScheduledJobs();
        setScheduledJobs(jobs);
    };

    const handleCancelJob = async (jobId: string) => {
        await cancelScheduledJob(jobId);
        fetchScheduledJobs(); // Refresh the list after cancellation
    };

    return (
        <Card style={{ padding: '20px' }}>
            <Heading size="4" style={{ marginBottom: '20px' }}>Scheduled Jobs</Heading>
            <Table.Root>
                <Table.Header>
                    <Table.Row>
                        <Table.ColumnHeaderCell>Job Name</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Start Time</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>CPU</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Memory</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Storage</Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell>Actions</Table.ColumnHeaderCell>
                    </Table.Row>
                </Table.Header>
                <Table.Body>
                    {scheduledJobs.map((scheduledJob) => (
                        <Table.Row key={scheduledJob.jobId}>
                            <Table.Cell>{job.name}</Table.Cell>
                            <Table.Cell>{format(new Date(scheduledJob.nextExecutionTime), 'yyyy-MM-dd HH:mm:ss')}</Table.Cell>
                            <Table.Cell>{scheduledJob.resourceRequirements.cpu}</Table.Cell>
                            <Table.Cell>{scheduledJob.resourceRequirements.memory}</Table.Cell>
                            <Table.Cell>{scheduledJob.resourceRequirements.storage}</Table.Cell>
                            <Table.Cell>
                                <Button size="1" color="red" onClick={() => handleCancelJob(scheduledJob.jobId)}>
                                    Cancel
                                </Button>
                            </Table.Cell>
                        </Table.Row>
                    ))}
                </Table.Body>
            </Table.Root>
            {scheduledJobs.length === 0 && (
                <Flex justify="center" style={{ marginTop: '20px' }}>
                    <Text>No scheduled jobs found.</Text>
                </Flex>
            )}
        </Card>
    );
};

export default ScheduledJobsList;
