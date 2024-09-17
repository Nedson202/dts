import React, { useState, useEffect } from 'react';
import { Calendar, dateFnsLocalizer } from 'react-big-calendar';
import { format, parse, startOfWeek, getDay } from 'date-fns';
import enUS from 'date-fns/locale/en-US';
import 'react-big-calendar/lib/css/react-big-calendar.css';
import { Heading, Card, Flex, Text } from '@radix-ui/themes';
import { Job, Resources, } from '../types';
import { getScheduledJobs } from '../services/jobService';

const locales = {
    'en-US': enUS,
};

const localizer = dateFnsLocalizer({
    format,
    parse,
    startOfWeek,
    getDay,
    locales,
});

interface CalendarEvent {
    id: string;
    title: string;
    start: Date;
    end: Date;
    resource: Resources;
}

const ScheduledJobsCalendar: React.FC<{ job: Job }> = ({ job }) => {
    const [events, setEvents] = useState<CalendarEvent[]>([]);

    useEffect(() => {
        const fetchScheduledJobs = async () => {
            const jobs = await getScheduledJobs();
            const calendarEvents = jobs.map(scheduledJob => ({
                id: job.id,
                title: job.name,
                start: new Date(scheduledJob.nextExecutionTime),
                end: new Date(scheduledJob.nextExecutionTime),
                resource: scheduledJob.resourceRequirements,
            }));
            setEvents(calendarEvents);
        };

        fetchScheduledJobs();
    }, []);

    const EventComponent: React.FC<{ event: CalendarEvent }> = ({ event }) => (
        <Flex direction="column">
            <Text weight="bold">{event.title}</Text>
            <Text size="1">CPU: {event.resource.cpu}</Text>
            <Text size="1">Memory: {event.resource.memory}</Text>
            <Text size="1">Storage: {event.resource.storage}</Text>
        </Flex>
    );

    return (
        <Card style={{ height: '500px', padding: '20px' }}>
            <Heading size="4" style={{ marginBottom: '20px' }}>Scheduled Jobs Calendar</Heading>
            <Calendar
                localizer={localizer}
                events={events}
                startAccessor="start"
                endAccessor="end"
                style={{ height: '100%' }}
                components={{
                    event: EventComponent,
                }}
            />
        </Card>
    );
};

export default ScheduledJobsCalendar;
