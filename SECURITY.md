# Security Policy

## Supported Versions

Chúng tôi hiện đang hỗ trợ các phiên bản sau đây với các bản cập nhật bảo mật:

| Version | Supported          |
| ------- | ------------------ |
| 2.x.x   | :white_check_mark: |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

### Báo cáo lỗ hổng bảo mật

**QUAN TRỌNG**: Không tạo public GitHub issue cho lỗ hổng bảo mật.

Nếu bạn phát hiện lỗ hổng bảo mật, vui lòng báo cáo một cách có trách nhiệm bằng cách:

#### Cách báo cáo ưu tiên
1. **GitHub Security Advisory**: Sử dụng tính năng "Security" > "Report a vulnerability" trên GitHub repository
2. **Email**: Gửi email đến tracepost.pro@gmail.com với thông tin chi tiết

#### Thông tin cần cung cấp
- Mô tả chi tiết về lỗ hổng
- Các bước để tái tạo lỗ hổng
- Phiên bản bị ảnh hưởng
- Potential impact và severity
- Proof of concept (nếu có)
- Đề xuất giải pháp (nếu có)

### Quy trình xử lý

1. **Xác nhận (24-48 giờ)**: Chúng tôi sẽ xác nhận đã nhận được báo cáo
2. **Đánh giá (3-5 ngày)**: Phân tích và đánh giá mức độ nghiêm trọng
3. **Sửa chữa (1-2 tuần)**: Phát triển và test patch
4. **Disclosure (sau khi fix)**: Công bố thông tin sau khi đã có patch

### Responsible Disclosure Timeline

- **90 ngày**: Thời gian tối đa để fix lỗ hổng critical
- **180 ngày**: Thời gian tối đa cho lỗ hổng ít nghiêm trọng hơn
- **Coordinated disclosure**: Sẽ thông báo trước khi public disclosure

## Security Measures

### Smart Contract Security

#### Auditing
- Smart contracts được audit bởi third-party security firms
- Regular security reviews cho major updates
- Bug bounty program cho việc tìm lỗ hổng

#### Best Practices
- Sử dụng OpenZeppelin contracts khi có thể
- Implement proper access controls
- Reentrancy protection
- Integer overflow/underflow protection
- Gas limit considerations

#### Testing
- Comprehensive unit tests
- Integration tests với realistic scenarios
- Formal verification cho critical functions
- Mainnet fork testing

### Backend Security

#### API Security
- Rate limiting
- Input validation và sanitization
- Authentication và authorization
- HTTPS only
- CORS configuration

#### Database Security
- Encrypted connections
- Regular backups với encryption
- Access controls
- SQL injection prevention

#### Infrastructure
- Regular security updates
- Monitoring và alerting
- WAF (Web Application Firewall)
- DDoS protection

### Frontend Security

#### Client-side Protection
- Content Security Policy (CSP)
- XSS prevention
- CSRF protection
- Secure cookie configuration

#### Wallet Integration
- Hardware wallet support
- Secure key management practices
- Transaction signing best practices
- Phishing protection

## Vulnerability Classes

### Critical (CVSS 9.0-10.0)
- Smart contract có thể bị exploit để steal funds
- Authentication bypass
- Remote code execution

### High (CVSS 7.0-8.9)
- Data breach possibilities
- Privilege escalation
- Significant financial impact

### Medium (CVSS 4.0-6.9)
- Information disclosure
- Denial of service
- Logic errors

### Low (CVSS 0.1-3.9)
- Minor information leaks
- Configuration issues
- Non-security impacting bugs

## Bug Bounty Program

### Scope
- Smart contracts trên mainnet
- API endpoints
- Frontend application
- Infrastructure components

### Rewards
- **Critical**: $5,000 - $50,000
- **High**: $1,000 - $5,000
- **Medium**: $500 - $1,000
- **Low**: $100 - $500

### Rules
- First come, first served
- Chỉ original discoveries
- No social engineering
- No DDoS attacks
- Must follow responsible disclosure

## Contact Information

### Security Team
- **Email**: tracepost.pro@gmail.com
- **PGP Key**: [Public key fingerprint]
- **Response time**: 24-48 hours

### Emergency Contact
Cho critical security issues:
- **Phone**: +[emergency-number]
- **Signal**: [secure-contact]

## Security Updates

### Notification Channels
- GitHub Security Advisories
- Project website security page
- Email notifications cho registered users
- Discord/Telegram announcements

### Update Process
1. Security patch development
2. Testing trên testnet
3. Community notification
4. Mainnet deployment
5. Post-deployment verification

## Acknowledgments

Chúng tôi ghi nhận và cảm ơn những security researchers đã giúp cải thiện bảo mật của dự án:

- [Researcher Name] - [Brief description of contribution]
- [Researcher Name] - [Brief description of contribution]

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Smart Contract Security Best Practices](https://consensys.github.io/smart-contract-best-practices/)
- [Ethereum Security Documentation](https://ethereum.org/en/developers/docs/security/)

---

**Lưu ý**: File này sẽ được cập nhật thường xuyên. Phiên bản mới nhất luôn có sẵn tại repository chính thức. 