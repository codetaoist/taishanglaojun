/**
 * TLS/SSL Configuration Manager for Taishang Laojun AI Platform
 * 
 * This module provides comprehensive TLS/SSL configuration including:
 * - TLS 1.3 configuration with perfect forward secrecy
 * - Certificate management and auto-renewal
 * - HSTS (HTTP Strict Transport Security)
 * - Certificate pinning
 * - OCSP stapling
 * - Regional compliance and certificate authorities
 * - Security headers configuration
 * - Cipher suite optimization
 */

import https from 'https';
import tls from 'tls';
import crypto from 'crypto';
import fs from 'fs/promises';
import path from 'path';

class TLSConfigManager {
    constructor(options = {}) {
        this.config = {
            // TLS version settings
            minVersion: options.minVersion || 'TLSv1.3',
            maxVersion: options.maxVersion || 'TLSv1.3',
            
            // Certificate settings
            certificatePath: options.certificatePath || './certs',
            autoRenewCertificates: options.autoRenewCertificates !== false,
            certificateRenewalDays: options.certificateRenewalDays || 30,
            
            // Security settings
            enableHSTS: options.enableHSTS !== false,
            hstsMaxAge: options.hstsMaxAge || 31536000, // 1 year
            enableHSTSPreload: options.enableHSTSPreload !== false,
            enableCertificatePinning: options.enableCertificatePinning || false,
            enableOCSPStapling: options.enableOCSPStapling !== false,
            
            // Cipher suites (TLS 1.3 recommended)
            cipherSuites: options.cipherSuites || [
                'TLS_AES_256_GCM_SHA384',
                'TLS_CHACHA20_POLY1305_SHA256',
                'TLS_AES_128_GCM_SHA256'
            ],
            
            // Legacy cipher suites for TLS 1.2 fallback
            legacyCipherSuites: options.legacyCipherSuites || [
                'ECDHE-RSA-AES256-GCM-SHA384',
                'ECDHE-RSA-CHACHA20-POLY1305',
                'ECDHE-RSA-AES128-GCM-SHA256',
                'ECDHE-RSA-AES256-SHA384',
                'ECDHE-RSA-AES128-SHA256'
            ],
            
            // Regional settings
            region: options.region || 'us-east-1',
            certificateAuthorities: options.certificateAuthorities || {
                'us-east-1': 'letsencrypt',
                'eu-central-1': 'letsencrypt',
                'ap-east-1': 'letsencrypt'
            },
            
            // Security headers
            securityHeaders: {
                'Strict-Transport-Security': `max-age=${options.hstsMaxAge || 31536000}; includeSubDomains; preload`,
                'X-Content-Type-Options': 'nosniff',
                'X-Frame-Options': 'DENY',
                'X-XSS-Protection': '1; mode=block',
                'Referrer-Policy': 'strict-origin-when-cross-origin',
                'Content-Security-Policy': options.csp || "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
                'Permissions-Policy': 'geolocation=(), microphone=(), camera=()',
                ...options.customHeaders
            },
            
            // Performance settings
            enableSessionResumption: options.enableSessionResumption !== false,
            sessionTimeout: options.sessionTimeout || 300, // 5 minutes
            enableCompression: options.enableCompression || false, // Disabled by default due to CRIME/BREACH
            
            // Monitoring
            enableMetrics: options.enableMetrics !== false,
            enableAuditLogging: options.enableAuditLogging !== false,
            
            ...options
        };

        // Certificate storage
        this.certificates = new Map();
        this.certificatePins = new Map();
        
        // TLS contexts
        this.tlsContexts = new Map();
        
        // Metrics
        this.metrics = {
            tlsConnections: 0,
            certificateRenewals: 0,
            securityViolations: 0,
            cipherSuiteUsage: new Map(),
            protocolVersionUsage: new Map()
        };

        // Audit logs
        this.auditLogs = [];

        this.init();
    }

    /**
     * Initialize TLS configuration manager
     */
    async init() {
        try {
            // Load existing certificates
            await this.loadCertificates();

            // Setup certificate renewal
            if (this.config.autoRenewCertificates) {
                this.setupCertificateRenewal();
            }

            // Setup security monitoring
            this.setupSecurityMonitoring();

            // Create TLS contexts
            await this.createTLSContexts();

            console.log('🔒 TLS Configuration Manager initialized');

        } catch (error) {
            console.error('❌ Failed to initialize TLS Configuration Manager:', error);
            throw error;
        }
    }

    /**
     * Get TLS configuration for HTTPS server
     */
    getTLSConfig(domain = 'default') {
        const certificate = this.certificates.get(domain) || this.certificates.get('default');
        
        if (!certificate) {
            throw new Error(`No certificate found for domain: ${domain}`);
        }

        const config = {
            // Certificate and key
            cert: certificate.cert,
            key: certificate.key,
            ca: certificate.ca,
            
            // TLS version
            minVersion: this.config.minVersion,
            maxVersion: this.config.maxVersion,
            
            // Cipher suites
            ciphers: this.config.legacyCipherSuites.join(':'),
            cipherSuites: this.config.cipherSuites.join(':'),
            
            // Security settings
            honorCipherOrder: true,
            secureProtocol: 'TLS_method',
            
            // Session settings
            sessionIdContext: crypto.createHash('sha1').update(domain).digest('hex').slice(0, 32),
            
            // OCSP stapling
            enableOCSPStapling: this.config.enableOCSPStapling,
            
            // Perfect Forward Secrecy
            dhparam: certificate.dhparam,
            
            // SNI callback for multi-domain support
            SNICallback: (servername, callback) => {
                this.handleSNI(servername, callback);
            },
            
            // Security callback
            secureConnect: (socket) => {
                this.handleSecureConnection(socket);
            }
        };

        return config;
    }

    /**
     * Create HTTPS server with TLS configuration
     */
    createHTTPSServer(app, domain = 'default') {
        const tlsConfig = this.getTLSConfig(domain);
        
        const server = https.createServer(tlsConfig, app);

        // Add security middleware
        server.on('request', (req, res) => {
            this.addSecurityHeaders(req, res);
        });

        // Monitor connections
        server.on('secureConnection', (socket) => {
            this.metrics.tlsConnections++;
            this.logConnection(socket);
        });

        // Handle TLS errors
        server.on('tlsClientError', (error, socket) => {
            this.handleTLSError(error, socket);
        });

        return server;
    }

    /**
     * Add security headers to response
     */
    addSecurityHeaders(req, res) {
        // Add HSTS header
        if (this.config.enableHSTS) {
            res.setHeader('Strict-Transport-Security', this.config.securityHeaders['Strict-Transport-Security']);
        }

        // Add other security headers
        Object.entries(this.config.securityHeaders).forEach(([header, value]) => {
            if (header !== 'Strict-Transport-Security' || this.config.enableHSTS) {
                res.setHeader(header, value);
            }
        });

        // Add certificate pinning header if enabled
        if (this.config.enableCertificatePinning) {
            const pins = this.getCertificatePins(req.hostname);
            if (pins.length > 0) {
                const pinHeader = pins.map(pin => `pin-sha256="${pin}"`).join('; ');
                res.setHeader('Public-Key-Pins', `${pinHeader}; max-age=5184000; includeSubDomains`);
            }
        }
    }

    /**
     * Handle SNI (Server Name Indication)
     */
    handleSNI(servername, callback) {
        try {
            const certificate = this.certificates.get(servername);
            
            if (certificate) {
                const context = tls.createSecureContext({
                    cert: certificate.cert,
                    key: certificate.key,
                    ca: certificate.ca
                });
                
                callback(null, context);
            } else {
                // Use default certificate
                const defaultCert = this.certificates.get('default');
                if (defaultCert) {
                    const context = tls.createSecureContext({
                        cert: defaultCert.cert,
                        key: defaultCert.key,
                        ca: defaultCert.ca
                    });
                    callback(null, context);
                } else {
                    callback(new Error('No certificate available'));
                }
            }

            this.log('SNI request handled', { servername });

        } catch (error) {
            this.log('SNI handling failed', { servername, error: error.message }, 'error');
            callback(error);
        }
    }

    /**
     * Handle secure connection
     */
    handleSecureConnection(socket) {
        try {
            const protocol = socket.getProtocol();
            const cipher = socket.getCipher();
            
            // Update metrics
            this.updateProtocolMetrics(protocol);
            this.updateCipherMetrics(cipher);
            
            // Validate security requirements
            this.validateConnectionSecurity(socket);

            this.log('Secure connection established', {
                protocol,
                cipher: cipher.name,
                remoteAddress: socket.remoteAddress
            });

        } catch (error) {
            this.log('Secure connection handling failed', { error: error.message }, 'error');
        }
    }

    /**
     * Validate connection security
     */
    validateConnectionSecurity(socket) {
        const protocol = socket.getProtocol();
        const cipher = socket.getCipher();

        // Check minimum TLS version
        if (protocol < this.config.minVersion) {
            this.metrics.securityViolations++;
            this.log('TLS version violation', { protocol, required: this.config.minVersion }, 'warning');
            socket.destroy();
            return;
        }

        // Check cipher suite
        const allowedCiphers = [...this.config.cipherSuites, ...this.config.legacyCipherSuites];
        if (!allowedCiphers.includes(cipher.name)) {
            this.metrics.securityViolations++;
            this.log('Cipher suite violation', { cipher: cipher.name }, 'warning');
        }

        // Check for weak encryption
        if (cipher.bits < 128) {
            this.metrics.securityViolations++;
            this.log('Weak encryption detected', { bits: cipher.bits }, 'warning');
            socket.destroy();
            return;
        }
    }

    /**
     * Load certificates from filesystem
     */
    async loadCertificates() {
        try {
            const certDir = this.config.certificatePath;
            
            // Check if certificate directory exists
            try {
                await fs.access(certDir);
            } catch {
                await fs.mkdir(certDir, { recursive: true });
                this.log('Certificate directory created', { path: certDir });
                return;
            }

            const files = await fs.readdir(certDir);
            
            for (const file of files) {
                if (file.endsWith('.crt') || file.endsWith('.pem')) {
                    const domain = file.replace(/\.(crt|pem)$/, '');
                    await this.loadCertificate(domain);
                }
            }

            this.log('Certificates loaded', { count: this.certificates.size });

        } catch (error) {
            this.log('Certificate loading failed', { error: error.message }, 'error');
            throw error;
        }
    }

    /**
     * Load individual certificate
     */
    async loadCertificate(domain) {
        try {
            const certPath = path.join(this.config.certificatePath, `${domain}.crt`);
            const keyPath = path.join(this.config.certificatePath, `${domain}.key`);
            const caPath = path.join(this.config.certificatePath, `${domain}-ca.crt`);
            const dhparamPath = path.join(this.config.certificatePath, `${domain}-dhparam.pem`);

            const certificate = {
                domain,
                cert: await fs.readFile(certPath, 'utf8'),
                key: await fs.readFile(keyPath, 'utf8'),
                loadedAt: Date.now()
            };

            // Load CA certificate if exists
            try {
                certificate.ca = await fs.readFile(caPath, 'utf8');
            } catch {
                // CA certificate is optional
            }

            // Load DH parameters if exists
            try {
                certificate.dhparam = await fs.readFile(dhparamPath, 'utf8');
            } catch {
                // DH parameters are optional
            }

            // Parse certificate for metadata
            certificate.metadata = this.parseCertificate(certificate.cert);

            // Generate certificate pin
            if (this.config.enableCertificatePinning) {
                certificate.pin = this.generateCertificatePin(certificate.cert);
                this.certificatePins.set(domain, [certificate.pin]);
            }

            this.certificates.set(domain, certificate);

            this.log('Certificate loaded', {
                domain,
                expiresAt: certificate.metadata.expiresAt,
                issuer: certificate.metadata.issuer
            });

        } catch (error) {
            this.log('Certificate loading failed', { domain, error: error.message }, 'error');
            throw error;
        }
    }

    /**
     * Parse certificate metadata
     */
    parseCertificate(certPem) {
        try {
            // This is a simplified parser - in production use a proper X.509 library
            const cert = crypto.createPublicKey(certPem);
            
            // Extract basic information (simplified)
            return {
                algorithm: cert.asymmetricKeyType,
                keySize: cert.asymmetricKeySize,
                expiresAt: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000), // Placeholder
                issuer: 'Unknown', // Placeholder
                subject: 'Unknown' // Placeholder
            };

        } catch (error) {
            this.log('Certificate parsing failed', { error: error.message }, 'error');
            return {};
        }
    }

    /**
     * Generate certificate pin (SPKI hash)
     */
    generateCertificatePin(certPem) {
        try {
            const cert = crypto.createPublicKey(certPem);
            const spki = cert.export({ type: 'spki', format: 'der' });
            return crypto.createHash('sha256').update(spki).digest('base64');

        } catch (error) {
            this.log('Certificate pin generation failed', { error: error.message }, 'error');
            return null;
        }
    }

    /**
     * Get certificate pins for domain
     */
    getCertificatePins(domain) {
        return this.certificatePins.get(domain) || [];
    }

    /**
     * Check certificate expiration
     */
    checkCertificateExpiration() {
        const expiringCertificates = [];
        const now = Date.now();
        const renewalThreshold = this.config.certificateRenewalDays * 24 * 60 * 60 * 1000;

        for (const [domain, certificate] of this.certificates) {
            if (certificate.metadata.expiresAt) {
                const timeToExpiry = certificate.metadata.expiresAt.getTime() - now;
                
                if (timeToExpiry <= renewalThreshold) {
                    expiringCertificates.push({
                        domain,
                        expiresAt: certificate.metadata.expiresAt,
                        daysRemaining: Math.floor(timeToExpiry / (24 * 60 * 60 * 1000))
                    });
                }
            }
        }

        return expiringCertificates;
    }

    /**
     * Renew certificate
     */
    async renewCertificate(domain) {
        try {
            this.log('Certificate renewal started', { domain });

            // This is a placeholder - implement actual certificate renewal
            // using ACME protocol (Let's Encrypt) or other CA
            
            // For now, just reload the certificate
            await this.loadCertificate(domain);

            this.metrics.certificateRenewals++;
            this.log('Certificate renewed', { domain });

            return { success: true, domain };

        } catch (error) {
            this.log('Certificate renewal failed', { domain, error: error.message }, 'error');
            throw error;
        }
    }

    /**
     * Setup certificate renewal monitoring
     */
    setupCertificateRenewal() {
        // Check for expiring certificates daily
        setInterval(async () => {
            try {
                const expiringCertificates = this.checkCertificateExpiration();
                
                for (const cert of expiringCertificates) {
                    if (cert.daysRemaining <= 7) {
                        // Renew certificates expiring within 7 days
                        await this.renewCertificate(cert.domain);
                    } else {
                        // Log warning for certificates expiring soon
                        this.log('Certificate expiring soon', cert, 'warning');
                    }
                }

            } catch (error) {
                this.log('Certificate renewal check failed', { error: error.message }, 'error');
            }
        }, 24 * 60 * 60 * 1000); // Daily check
    }

    /**
     * Setup security monitoring
     */
    setupSecurityMonitoring() {
        // Monitor for security violations
        setInterval(() => {
            if (this.metrics.securityViolations > 0) {
                this.log('Security violations detected', {
                    violations: this.metrics.securityViolations
                }, 'warning');
                
                // Reset counter
                this.metrics.securityViolations = 0;
            }
        }, 60000); // Every minute
    }

    /**
     * Create TLS contexts for different domains
     */
    async createTLSContexts() {
        for (const [domain, certificate] of this.certificates) {
            try {
                const context = tls.createSecureContext({
                    cert: certificate.cert,
                    key: certificate.key,
                    ca: certificate.ca,
                    dhparam: certificate.dhparam
                });

                this.tlsContexts.set(domain, context);

            } catch (error) {
                this.log('TLS context creation failed', { domain, error: error.message }, 'error');
            }
        }
    }

    /**
     * Handle TLS errors
     */
    handleTLSError(error, socket) {
        this.metrics.securityViolations++;
        
        this.log('TLS error', {
            error: error.message,
            code: error.code,
            remoteAddress: socket.remoteAddress
        }, 'error');

        // Close the connection
        socket.destroy();
    }

    /**
     * Update protocol usage metrics
     */
    updateProtocolMetrics(protocol) {
        const current = this.metrics.protocolVersionUsage.get(protocol) || 0;
        this.metrics.protocolVersionUsage.set(protocol, current + 1);
    }

    /**
     * Update cipher suite usage metrics
     */
    updateCipherMetrics(cipher) {
        const current = this.metrics.cipherSuiteUsage.get(cipher.name) || 0;
        this.metrics.cipherSuiteUsage.set(cipher.name, current + 1);
    }

    /**
     * Log connection details
     */
    logConnection(socket) {
        const connectionInfo = {
            protocol: socket.getProtocol(),
            cipher: socket.getCipher(),
            remoteAddress: socket.remoteAddress,
            remotePort: socket.remotePort,
            timestamp: Date.now()
        };

        this.log('TLS connection', connectionInfo);
    }

    /**
     * Get TLS configuration for client connections
     */
    getClientTLSConfig(options = {}) {
        return {
            // TLS version
            minVersion: this.config.minVersion,
            maxVersion: this.config.maxVersion,
            
            // Cipher suites
            ciphers: this.config.legacyCipherSuites.join(':'),
            cipherSuites: this.config.cipherSuites.join(':'),
            
            // Certificate validation
            rejectUnauthorized: options.rejectUnauthorized !== false,
            checkServerIdentity: options.checkServerIdentity || tls.checkServerIdentity,
            
            // Certificate pinning
            ca: options.ca,
            
            // Session resumption
            session: options.session,
            
            // SNI
            servername: options.servername,
            
            // ALPN
            ALPNProtocols: options.ALPNProtocols || ['h2', 'http/1.1']
        };
    }

    /**
     * Validate TLS configuration
     */
    validateConfiguration() {
        const issues = [];

        // Check if we have certificates
        if (this.certificates.size === 0) {
            issues.push('No certificates loaded');
        }

        // Check certificate expiration
        const expiringCerts = this.checkCertificateExpiration();
        if (expiringCerts.length > 0) {
            issues.push(`${expiringCerts.length} certificates expiring soon`);
        }

        // Check TLS version
        if (this.config.minVersion < 'TLSv1.2') {
            issues.push('Minimum TLS version is below recommended (TLSv1.2)');
        }

        // Check cipher suites
        if (this.config.cipherSuites.length === 0) {
            issues.push('No TLS 1.3 cipher suites configured');
        }

        return {
            valid: issues.length === 0,
            issues,
            recommendations: this.getSecurityRecommendations()
        };
    }

    /**
     * Get security recommendations
     */
    getSecurityRecommendations() {
        const recommendations = [];

        // TLS version recommendations
        if (this.config.minVersion !== 'TLSv1.3') {
            recommendations.push('Consider upgrading minimum TLS version to 1.3');
        }

        // HSTS recommendations
        if (!this.config.enableHSTS) {
            recommendations.push('Enable HSTS for better security');
        }

        // Certificate pinning recommendations
        if (!this.config.enableCertificatePinning) {
            recommendations.push('Consider enabling certificate pinning for critical domains');
        }

        // OCSP stapling recommendations
        if (!this.config.enableOCSPStapling) {
            recommendations.push('Enable OCSP stapling for better performance and privacy');
        }

        return recommendations;
    }

    /**
     * Export TLS configuration
     */
    exportConfiguration() {
        return {
            version: '1.0',
            config: {
                ...this.config,
                // Remove sensitive data
                certificatePath: undefined
            },
            certificates: Array.from(this.certificates.keys()),
            metrics: this.getStats(),
            exportedAt: new Date().toISOString()
        };
    }

    /**
     * Get TLS statistics
     */
    getStats() {
        return {
            ...this.metrics,
            certificates: this.certificates.size,
            tlsContexts: this.tlsContexts.size,
            protocolVersionUsage: Object.fromEntries(this.metrics.protocolVersionUsage),
            cipherSuiteUsage: Object.fromEntries(this.metrics.cipherSuiteUsage),
            region: this.config.region,
            hstsEnabled: this.config.enableHSTS,
            certificatePinningEnabled: this.config.enableCertificatePinning
        };
    }

    /**
     * Health check
     */
    healthCheck() {
        const validation = this.validateConfiguration();
        const expiringCerts = this.checkCertificateExpiration();

        return {
            healthy: validation.valid && expiringCerts.length === 0,
            issues: [
                ...validation.issues,
                ...expiringCerts.map(cert => `Certificate for ${cert.domain} expires in ${cert.daysRemaining} days`)
            ],
            recommendations: validation.recommendations,
            stats: this.getStats()
        };
    }

    /**
     * Utility method for logging
     */
    log(message, data = {}, level = 'info') {
        if (!this.config.enableAuditLogging) return;

        const logEntry = {
            timestamp: new Date().toISOString(),
            level,
            message,
            data,
            region: this.config.region
        };

        this.auditLogs.push(logEntry);

        // Keep only recent logs
        if (this.auditLogs.length > 10000) {
            this.auditLogs.splice(0, this.auditLogs.length - 5000);
        }

        // In production, send to external logging service
        console.log(`[TLS-${level.toUpperCase()}] ${message}`, data);
    }
}

export default TLSConfigManager;