import React, { useEffect } from 'react';
import {
    Grid,
    Paper,
    Typography,
    Box,
    Card,
    CardContent,
    List,
    ListItem,
    ListItemText,
    ListItemSecondaryAction,
    Chip,
    IconButton,
    Tooltip,
} from '@mui/material';
import {
    TrendingUp,
    TrendingDown,
    Phone,
    CheckCircle,
    Error,
    Refresh,
    AccountTree as AccountTreeIcon,
} from '@mui/icons-material';
import { observer } from 'mobx-react-lite';
import { useStores } from '../../hooks/useStores';
import {
    AreaChart,
    Area,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip as RechartsTooltip,
    ResponsiveContainer,
    PieChart,
    Pie,
    Cell,
} from 'recharts';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

export const Dashboard: React.FC = observer(() => {
    const { flowStore, dataStore } = useStores();

    useEffect(() => {
        // Load dashboard data
        flowStore.loadFlows();
        dataStore.loadAllData();
    }, [flowStore, dataStore]);

    // Mock data for charts
    const dailyStats = [
        { date: 'Пн', leads: 65, calls: 45, success: 30 },
        { date: 'Вт', leads: 78, calls: 52, success: 35 },
        { date: 'Ср', leads: 82, calls: 58, success: 42 },
        { date: 'Чт', leads: 73, calls: 48, success: 32 },
        { date: 'Пт', leads: 89, calls: 62, success: 45 },
        { date: 'Сб', leads: 45, calls: 30, success: 20 },
        { date: 'Вс', leads: 38, calls: 25, success: 18 },
    ];

    const conversionData = [
        { name: 'Успешные', value: 342 },
        { name: 'В работе', value: 189 },
        { name: 'Отклонены', value: 87 },
        { name: 'Нет ответа', value: 156 },
    ];

    const recentActivities = [
        { id: 1, type: 'success', text: 'Сделка #1234 отправлена в автообзвон', time: '5 мин назад' },
        { id: 2, type: 'info', text: 'Обновлен поток "Холодные звонки"', time: '15 мин назад' },
        { id: 3, type: 'warning', text: 'Ошибка при обработке сделки #5678', time: '1 час назад' },
        { id: 4, type: 'success', text: 'Синхронизация с AmoCRM завершена', time: '2 часа назад' },
    ];

    const StatCard: React.FC<{
        title: string;
        value: string | number;
        change?: number;
        icon: React.ReactNode;
        color?: string;
    }> = ({ title, value, change, icon, color = 'primary' }) => (
        <Card>
            <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <Box>
                        <Typography color="textSecondary" gutterBottom>
                            {title}
                        </Typography>
                        <Typography variant="h4">{value}</Typography>
                        {change !== undefined && (
                            <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
                                {change > 0 ? (
                                    <TrendingUp color="success" fontSize="small" />
                                ) : (
                                    <TrendingDown color="error" fontSize="small" />
                                )}
                                <Typography
                                    variant="body2"
                                    color={change > 0 ? 'success.main' : 'error.main'}
                                    sx={{ ml: 0.5 }}
                                >
                                    {Math.abs(change)}%
                                </Typography>
                            </Box>
                        )}
                    </Box>
                    <Box
                        sx={{
                            backgroundColor: `${color}.light`,
                            borderRadius: '50%',
                            p: 2,
                            color: `${color}.main`,
                        }}
                    >
                        {icon}
                    </Box>
                </Box>
            </CardContent>
        </Card>
    );

    return (
        <Box>
            <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="h4">Дашборд</Typography>
                <Tooltip title="Обновить данные">
                    <IconButton>
                        <Refresh />
                    </IconButton>
                </Tooltip>
            </Box>

            <Grid container spacing={3}>
                {/* Stat Cards */}
                <Grid item xs={12} sm={6} md={3}>
                    <StatCard
                        title="Активные потоки"
                        value={flowStore.flows.filter(f => f.is_active).length}
                        icon={<AccountTreeIcon />}
                        color="primary"
                    />
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                    <StatCard
                        title="Сделки за сегодня"
                        value="142"
                        change={12}
                        icon={<TrendingUp />}
                        color="success"
                    />
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                    <StatCard
                        title="Звонки за сегодня"
                        value="89"
                        change={-5}
                        icon={<Phone />}
                        color="info"
                    />
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                    <StatCard
                        title="Конверсия"
                        value="34%"
                        change={3}
                        icon={<CheckCircle />}
                        color="warning"
                    />
                </Grid>

                {/* Daily Statistics Chart */}
                <Grid item xs={12} md={8}>
                    <Paper sx={{ p: 2 }}>
                        <Typography variant="h6" gutterBottom>
                            Статистика за неделю
                        </Typography>
                        <ResponsiveContainer width="100%" height={300}>
                            <AreaChart data={dailyStats}>
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="date" />
                                <YAxis />
                                <RechartsTooltip />
                                <Area
                                    type="monotone"
                                    dataKey="leads"
                                    stackId="1"
                                    stroke="#8884d8"
                                    fill="#8884d8"
                                    name="Сделки"
                                />
                                <Area
                                    type="monotone"
                                    dataKey="calls"
                                    stackId="1"
                                    stroke="#82ca9d"
                                    fill="#82ca9d"
                                    name="Звонки"
                                />
                                <Area
                                    type="monotone"
                                    dataKey="success"
                                    stackId="1"
                                    stroke="#ffc658"
                                    fill="#ffc658"
                                    name="Успешные"
                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    </Paper>
                </Grid>

                {/* Conversion Pie Chart */}
                <Grid item xs={12} md={4}>
                    <Paper sx={{ p: 2, height: '100%' }}>
                        <Typography variant="h6" gutterBottom>
                            Распределение сделок
                        </Typography>
                        <ResponsiveContainer width="100%" height={300}>
                            <PieChart>
                                <Pie
                                    data={conversionData}
                                    cx="50%"
                                    cy="50%"
                                    labelLine={false}
                                    label={(entry) => `${entry.name}: ${entry.value}`}
                                    outerRadius={80}
                                    fill="#8884d8"
                                    dataKey="value"
                                >
                                    {conversionData.map((entry, index) => (
                                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                    ))}
                                </Pie>
                                <RechartsTooltip />
                            </PieChart>
                        </ResponsiveContainer>
                    </Paper>
                </Grid>

                {/* Recent Activities */}
                <Grid item xs={12} md={6}>
                    <Paper sx={{ p: 2 }}>
                        <Typography variant="h6" gutterBottom>
                            Последние действия
                        </Typography>
                        <List>
                            {recentActivities.map((activity) => (
                                <ListItem key={activity.id}>
                                    <ListItemText
                                        primary={activity.text}
                                        secondary={activity.time}
                                    />
                                    <ListItemSecondaryAction>
                                        <Chip
                                            label={activity.type}
                                            size="small"
                                            color={
                                                activity.type === 'success'
                                                    ? 'success'
                                                    : activity.type === 'warning'
                                                        ? 'warning'
                                                        : 'info'
                                            }
                                        />
                                    </ListItemSecondaryAction>
                                </ListItem>
                            ))}
                        </List>
                    </Paper>
                </Grid>

                {/* System Status */}
                <Grid item xs={12} md={6}>
                    <Paper sx={{ p: 2 }}>
                        <Typography variant="h6" gutterBottom>
                            Статус системы
                        </Typography>
                        <List>
                            <ListItem>
                                <ListItemText
                                    primary="AmoCRM API"
                                    secondary="Подключено, последняя синхронизация: 10 мин назад"
                                />
                                <ListItemSecondaryAction>
                                    <Chip icon={<CheckCircle />} label="OK" color="success" size="small" />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <ListItem>
                                <ListItemText
                                    primary="Система автообзвона"
                                    secondary="Активно, очередь: 23 контакта"
                                />
                                <ListItemSecondaryAction>
                                    <Chip icon={<CheckCircle />} label="OK" color="success" size="small" />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <ListItem>
                                <ListItemText
                                    primary="Обработка вебхуков"
                                    secondary="5 вебхуков в очереди"
                                />
                                <ListItemSecondaryAction>
                                    <Chip icon={<Error />} label="Warning" color="warning" size="small" />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <ListItem>
                                <ListItemText
                                    primary="База данных"
                                    secondary="PostgreSQL: 15% использования"
                                />
                                <ListItemSecondaryAction>
                                    <Chip icon={<CheckCircle />} label="OK" color="success" size="small" />
                                </ListItemSecondaryAction>
                            </ListItem>
                        </List>
                    </Paper>
                </Grid>
            </Grid>
        </Box>
    );
});