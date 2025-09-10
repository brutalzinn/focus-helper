#!/usr/bin/env node

/**
 * Focus Helper MCP Server for Cursor
 * This server bridges Cursor with the Focus Helper application
 */

const http = require('http');
const readline = require('readline');

class FocusHelperMCPServer {
    constructor() {
        this.baseUrl = 'http://localhost:8089';
        this.rl = readline.createInterface({
            input: process.stdin,
            output: process.stdout
        });
        
        this.setupHandlers();
    }

    setupHandlers() {
        this.rl.on('line', (line) => {
            try {
                const request = JSON.parse(line);
                this.handleRequest(request);
            } catch (error) {
                this.sendError(null, -32700, 'Parse error', error.message);
            }
        });
    }

    async handleRequest(request) {
        try {
            const { method, params } = request;

            switch (method) {
                case 'initialize':
                    this.sendResponse(request.id, {
                        protocolVersion: '2024-11-05',
                        capabilities: {
                            tools: {}
                        },
                        serverInfo: {
                            name: 'focus-helper-mcp',
                            version: '1.0.0'
                        }
                    });
                    break;

                case 'tools/list':
                    this.sendResponse(request.id, {
                        tools: [
                            {
                                name: 'get_session_info',
                                description: 'Get current focus session information including duration, hyperfocus status, and activity',
                                inputSchema: {
                                    type: 'object',
                                    properties: {}
                                }
                            },
                            {
                                name: 'get_alert_levels',
                                description: 'Get available alert levels and their thresholds',
                                inputSchema: {
                                    type: 'object',
                                    properties: {}
                                }
                            },
                            {
                                name: 'trigger_alert',
                                description: 'Trigger a specific alert level by index',
                                inputSchema: {
                                    type: 'object',
                                    properties: {
                                        alert_index: {
                                            type: 'number',
                                            description: 'Index of the alert level to trigger'
                                        }
                                    },
                                    required: ['alert_index']
                                }
                            },
                            {
                                name: 'get_hyperfocus_status',
                                description: 'Get current hyperfocus status and duration',
                                inputSchema: {
                                    type: 'object',
                                    properties: {}
                                }
                            },
                            {
                                name: 'ping',
                                description: 'Test connection to the focus helper MCP server',
                                inputSchema: {
                                    type: 'object',
                                    properties: {}
                                }
                            }
                        ]
                    });
                    break;

                case 'tools/call':
                    await this.handleToolCall(request);
                    break;

                case 'ping':
                    // Handle direct ping requests
                    try {
                        const result = await this.sendMCPRequest('ping', {});
                        this.sendResponse(request.id, result);
                    } catch (error) {
                        this.sendError(request.id, -32603, 'Ping failed', error.message);
                    }
                    break;

                default:
                    this.sendError(request.id, -32601, 'Method not found');
            }
        } catch (error) {
            this.sendError(request.id, -32603, 'Internal error', error.message);
        }
    }

    async handleToolCall(request) {
        const { name, arguments: args } = request.params;

        try {
            let result;
            switch (name) {
                case 'get_session_info':
                    result = await this.sendMCPRequest('get_session_info', {});
                    break;
                case 'get_alert_levels':
                    result = await this.sendMCPRequest('get_alert_levels', {});
                    break;
                case 'trigger_alert':
                    result = await this.sendMCPRequest('trigger_alert', args);
                    break;
                case 'get_hyperfocus_status':
                    result = await this.sendMCPRequest('get_hyperfocus_status', {});
                    break;
                case 'ping':
                    result = await this.sendMCPRequest('ping', {});
                    break;
                default:
                    throw new Error(`Unknown tool: ${name}`);
            }

            this.sendResponse(request.id, {
                content: [
                    {
                        type: 'text',
                        text: JSON.stringify(result, null, 2)
                    }
                ]
            });
        } catch (error) {
            this.sendError(request.id, -32603, 'Tool execution failed', error.message);
        }
    }

    async sendMCPRequest(method, params) {
        return new Promise((resolve, reject) => {
            const data = JSON.stringify({
                jsonrpc: '2.0',
                id: Date.now(),
                method,
                params
            });

            const options = {
                hostname: 'localhost',
                port: 8089,
                path: '/mcp',
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Content-Length': Buffer.byteLength(data)
                }
            };

            const req = http.request(options, (res) => {
                let body = '';
                res.on('data', chunk => body += chunk);
                res.on('end', () => {
                    try {
                        const response = JSON.parse(body);
                        if (response.error) {
                            reject(new Error(response.error.message));
                        } else {
                            resolve(response.result);
                        }
                    } catch (e) {
                        reject(e);
                    }
                });
            });

            req.on('error', reject);
            req.write(data);
            req.end();
        });
    }

    sendResponse(id, result) {
        console.log(JSON.stringify({
            jsonrpc: '2.0',
            id,
            result
        }));
    }

    sendError(id, code, message, data = null) {
        console.log(JSON.stringify({
            jsonrpc: '2.0',
            id,
            error: {
                code,
                message,
                data
            }
        }));
    }
}

// Start the server
new FocusHelperMCPServer();
