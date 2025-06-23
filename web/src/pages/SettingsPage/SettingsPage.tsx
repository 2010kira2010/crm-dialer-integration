import React, { useState } from 'react';
import {
    Chip,
    Box,
    Typography,
    Paper,
    Tabs,
    Tab,
    Button,
    TextField,
    Switch,
    FormControlLabel,
    List,
    ListItem,
    ListItemText,
    ListItemSecondaryAction,
    Alert,
    CircularProgress,
    Divider,
    Grid,
    Card,
    CardContent,
    CardActions,
} from '@mui/material';
import {
    Refresh as RefreshIcon,
    Check as CheckIcon,
    Error as ErrorIcon,
    Link as LinkIcon,
    LinkOff as LinkOffIcon,
} from '@mui/icons-material';
import { observer } from 'mobx-react-lite';
import { useStores } from '../../hooks/useStores';

interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
}

function TabPanel(props: TabPanelProps) {
    const { children, value, index, ...other } = props;
    return (
        <div
            role="tabpanel"
            hidden={value !== index}
            id={`settings-tabpanel-${index}`}
            aria-labelledby={`settings-tab-${index}`}
            {...other}
        >
            {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
        </div>
    );
}

export const SettingsPage: React.FC = observer(() => {
    const { dataStore } = useStores();
    const [tabValue, setTabValue] = useState(0);
    const [amocrmConnected, setAmocrmConnected] = useState(false);
    const [dialerConnected, setDialerConnected] = useState(true);
    const [isSyncing, setIsSyncing] = useState(false);

    const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
        setTabValue(newValue);
    };

    const handleAmoCRMConnect = () => {
        // Open AmoCRM OAuth window
        const authUrl = `/api/v1/amocrm/auth?state=${Date.now()}`;
        window.open(authUrl, 'amocrm-auth', 'width=600,height=700');
    };

    const handleSyncFields = async () => {
        setIsSyncing(true);
        try {
            // TODO: Call API to sync fields
            await new Promise(resolve => setTimeout(resolve, 2000));
            await dataStore.loadAllData();
        } finally {
            setIsSyncing(false);
        }
    };

    const handleSyncDialerData = async () => {
        setIsSyncing(true);
        try {
            // TODO: Call API to sync dialer data
            await new Promise(resolve => setTimeout(resolve, 2000));
            await dataStore.loadAllData();
        } finally {
            setIsSyncing(false);
        }
    };

    return (
        <Box>
            <Typography variant="h4" gutterBottom>
                Настройки
            </Typography>

            <Paper sx={{ width: '100%' }}>
                <Tabs value={tabValue} onChange={handleTabChange} aria-label="settings tabs">
                    <Tab label="Интеграции" />
                    <Tab label="Справочники" />
                    <Tab label="Вебхуки" />
                    <Tab label="Общие" />
                </Tabs>

                <TabPanel value={tabValue} index={0}>
                    <Grid container spacing={3}>
                        {/* AmoCRM Integration */}
                        <Grid item xs={12} md={6}>
                            <Card>
                                <CardContent>
                                    <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                                        <Typography variant="h6" sx={{ flexGrow: 1 }}>
                                            AmoCRM
                                        </Typography>
                                        {amocrmConnected ? (
                                            <Chip
                                                icon={<CheckIcon />}
                                                label="Подключено"
                                                color="success"
                                                size="small"
                                            />
                                        ) : (
                                            <Chip
                                                icon={<ErrorIcon />}
                                                label="Не подключено"
                                                color="error"
                                                size="small"
                                            />
                                        )}
                                    </Box>

                                    <Typography variant="body2" color="text.secondary" gutterBottom>
                                        Интеграция с AmoCRM для получения сделок и контактов
                                    </Typography>

                                    {amocrmConnected && (
                                        <Box sx={{ mt: 2 }}>
                                            <Typography variant="body2">
                                                Домен: your-domain.amocrm.ru
                                            </Typography>
                                            <Typography variant="body2">
                                                Последняя синхронизация: 10 минут назад
                                            </Typography>
                                        </Box>
                                    )}
                                </CardContent>
                                <CardActions>
                                    {amocrmConnected ? (
                                        <>
                                            <Button size="small" startIcon={<RefreshIcon />} onClick={handleSyncFields}>
                                                Синхронизировать
                                            </Button>
                                            <Button size="small" color="error" startIcon={<LinkOffIcon />}>
                                                Отключить
                                            </Button>
                                        </>
                                    ) : (
                                        <Button
                                            variant="contained"
                                            size="small"
                                            startIcon={<LinkIcon />}
                                            onClick={handleAmoCRMConnect}
                                        >
                                            Подключить
                                        </Button>
                                    )}
                                </CardActions>
                            </Card>
                        </Grid>

                        {/* Dialer Integration */}
                        <Grid item xs={12} md={6}>
                            <Card>
                                <CardContent>
                                    <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                                        <Typography variant="h6" sx={{ flexGrow: 1 }}>
                                            Система автообзвона
                                        </Typography>
                                        {dialerConnected ? (
                                            <Chip
                                                icon={<CheckIcon />}
                                                label="Подключено"
                                                color="success"
                                                size="small"
                                            />
                                        ) : (
                                            <Chip
                                                icon={<ErrorIcon />}
                                                label="Не подключено"
                                                color="error"
                                                size="small"
                                            />
                                        )}
                                    </Box>

                                    <Typography variant="body2" color="text.secondary" gutterBottom>
                                        Интеграция с системой автообзвона для передачи контактов
                                    </Typography>

                                    {dialerConnected && (
                                        <Box sx={{ mt: 2 }}>
                                            <Typography variant="body2">
                                                API URL: https://dialer-api.example.com
                                            </Typography>
                                            <Typography variant="body2">
                                                Активных кампаний: {dataStore.dialerCampaigns.length}
                                            </Typography>
                                        </Box>
                                    )}
                                </CardContent>
                                <CardActions>
                                    {dialerConnected ? (
                                        <>
                                            <Button
                                                size="small"
                                                startIcon={<RefreshIcon />}
                                                onClick={handleSyncDialerData}
                                            >
                                                Синхронизировать
                                            </Button>
                                            <Button size="small" color="error" startIcon={<LinkOffIcon />}>
                                                Отключить
                                            </Button>
                                        </>
                                    ) : (
                                        <Button variant="contained" size="small" startIcon={<LinkIcon />}>
                                            Подключить
                                        </Button>
                                    )}
                                </CardActions>
                            </Card>
                        </Grid>
                    </Grid>
                </TabPanel>

                <TabPanel value={tabValue} index={1}>
                    <Box sx={{ mb: 3 }}>
                        <Typography variant="h6" gutterBottom>
                            Справочники данных
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                            Синхронизация справочников из подключенных систем
                        </Typography>
                    </Box>

                    {isSyncing ? (
                        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                            <CircularProgress />
                        </Box>
                    ) : (
                        <Grid container spacing={3}>
                            <Grid item xs={12} md={6}>
                                <Paper sx={{ p: 2 }}>
                                    <Typography variant="subtitle1" gutterBottom>
                                        Поля AmoCRM
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary" gutterBottom>
                                        Всего полей: {dataStore.amocrmFields.length}
                                    </Typography>
                                    <List dense>
                                        {dataStore.amocrmFields.slice(0, 5).map((field) => (
                                            <ListItem key={field.id}>
                                                <ListItemText
                                                    primary={field.name}
                                                    secondary={`Тип: ${field.type}`}
                                                />
                                            </ListItem>
                                        ))}
                                    </List>
                                    <Button
                                        fullWidth
                                        variant="outlined"
                                        startIcon={<RefreshIcon />}
                                        onClick={handleSyncFields}
                                        disabled={isSyncing}
                                    >
                                        Синхронизировать поля
                                    </Button>
                                </Paper>
                            </Grid>

                            <Grid item xs={12} md={6}>
                                <Paper sx={{ p: 2 }}>
                                    <Typography variant="subtitle1" gutterBottom>
                                        Кампании автообзвона
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary" gutterBottom>
                                        Всего кампаний: {dataStore.dialerCampaigns.length}
                                    </Typography>
                                    <List dense>
                                        {dataStore.dialerCampaigns.slice(0, 5).map((campaign) => (
                                            <ListItem key={campaign.id}>
                                                <ListItemText primary={campaign.name} />
                                            </ListItem>
                                        ))}
                                    </List>
                                    <Button
                                        fullWidth
                                        variant="outlined"
                                        startIcon={<RefreshIcon />}
                                        onClick={handleSyncDialerData}
                                        disabled={isSyncing}
                                    >
                                        Синхронизировать кампании
                                    </Button>
                                </Paper>
                            </Grid>
                        </Grid>
                    )}
                </TabPanel>

                <TabPanel value={tabValue} index={2}>
                    <Box sx={{ mb: 3 }}>
                        <Typography variant="h6" gutterBottom>
                            Настройка вебхуков
                        </Typography>
                        <Alert severity="info" sx={{ mb: 2 }}>
                            URL для вебхуков: https://your-domain.com/api/v1/webhooks/amocrm
                        </Alert>
                    </Box>

                    <List>
                        <ListItem>
                            <ListItemText
                                primary="Создание сделки"
                                secondary="Срабатывает при создании новой сделки в AmoCRM"
                            />
                            <ListItemSecondaryAction>
                                <Switch defaultChecked />
                            </ListItemSecondaryAction>
                        </ListItem>
                        <Divider />
                        <ListItem>
                            <ListItemText
                                primary="Обновление сделки"
                                secondary="Срабатывает при изменении любых данных сделки"
                            />
                            <ListItemSecondaryAction>
                                <Switch defaultChecked />
                            </ListItemSecondaryAction>
                        </ListItem>
                        <Divider />
                        <ListItem>
                            <ListItemText
                                primary="Смена статуса"
                                secondary="Срабатывает при изменении статуса сделки"
                            />
                            <ListItemSecondaryAction>
                                <Switch defaultChecked />
                            </ListItemSecondaryAction>
                        </ListItem>
                        <Divider />
                        <ListItem>
                            <ListItemText
                                primary="Смена ответственного"
                                secondary="Срабатывает при изменении ответственного за сделку"
                            />
                            <ListItemSecondaryAction>
                                <Switch defaultChecked />
                            </ListItemSecondaryAction>
                        </ListItem>
                    </List>
                </TabPanel>

                <TabPanel value={tabValue} index={3}>
                    <Grid container spacing={3}>
                        <Grid item xs={12} md={6}>
                            <Typography variant="h6" gutterBottom>
                                Общие настройки
                            </Typography>

                            <Box sx={{ mt: 3 }}>
                                <TextField
                                    fullWidth
                                    label="Название компании"
                                    defaultValue="Your Company"
                                    margin="normal"
                                />

                                <TextField
                                    fullWidth
                                    label="Email для уведомлений"
                                    type="email"
                                    defaultValue="admin@example.com"
                                    margin="normal"
                                />

                                <FormControlLabel
                                    control={<Switch defaultChecked />}
                                    label="Включить email уведомления"
                                    sx={{ mt: 2 }}
                                />

                                <FormControlLabel
                                    control={<Switch defaultChecked />}
                                    label="Логировать все действия"
                                    sx={{ mt: 1 }}
                                />
                            </Box>
                        </Grid>

                        <Grid item xs={12} md={6}>
                            <Typography variant="h6" gutterBottom>
                                Лимиты и ограничения
                            </Typography>

                            <Box sx={{ mt: 3 }}>
                                <TextField
                                    fullWidth
                                    label="Макс. запросов к AmoCRM в секунду"
                                    type="number"
                                    defaultValue="7"
                                    margin="normal"
                                    helperText="Рекомендуемое значение: 7"
                                />

                                <TextField
                                    fullWidth
                                    label="Макс. сущностей в батче"
                                    type="number"
                                    defaultValue="200"
                                    margin="normal"
                                    helperText="Максимум: 200"
                                />

                                <TextField
                                    fullWidth
                                    label="Таймаут запросов (сек)"
                                    type="number"
                                    defaultValue="30"
                                    margin="normal"
                                />
                            </Box>
                        </Grid>
                    </Grid>

                    <Box sx={{ mt: 4, display: 'flex', justifyContent: 'flex-end' }}>
                        <Button variant="contained" size="large">
                            Сохранить настройки
                        </Button>
                    </Box>
                </TabPanel>
            </Paper>
        </Box>
    );
});