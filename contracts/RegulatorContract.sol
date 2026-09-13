// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./TraceEvidence.sol";

/**
 * @title RegulatorContract
 * @dev 监管合约，用于监管部门查看审计数据
 */
contract RegulatorContract {
    
    TraceEvidence public traceEvidence;
    
    // 监管者地址映射
    mapping(address => bool) public regulators;
    
    // 审计日志结构
    struct AuditLog {
        string traceID;
        string operation;
        address regulator;
        uint256 timestamp;
        string details;
    }
    
    AuditLog[] public auditLogs;
    
    // 事件
    event AuditPerformed(
        string indexed traceID,
        address indexed regulator,
        string operation,
        uint256 timestamp
    );
    
    modifier onlyRegulator() {
        require(regulators[msg.sender], "Only regulator can perform this action");
        _;
    }
    
    /**
     * @dev 构造函数
     * @param _traceEvidenceAddress 存证合约地址
     */
    constructor(address _traceEvidenceAddress) {
        traceEvidence = TraceEvidence(_traceEvidenceAddress);
        regulators[msg.sender] = true; // 部署者自动成为监管者
    }
    
    /**
     * @dev 添加监管者
     * @param _regulator 监管者地址
     */
    function addRegulator(address _regulator) public onlyRegulator {
        regulators[_regulator] = true;
    }
    
    /**
     * @dev 移除监管者
     * @param _regulator 监管者地址
     */
    function removeRegulator(address _regulator) public onlyRegulator {
        regulators[_regulator] = false;
    }
    
    /**
     * @dev 查询所有操作日志
     * @param _traceID 追溯码
     * @return 操作历史记录数组
     */
    function queryAllOperations(string memory _traceID)
        public
        view
        onlyRegulator
        returns (
            string[] memory operationTypes,
            string[] memory operators,
            uint256[] memory timestamps
        )
    {
        uint256 count = traceEvidence.getHistoryCount(_traceID);
        operationTypes = new string[](count);
        operators = new string[](count);
        timestamps = new uint256[](count);
        
        for (uint256 i = 0; i < count; i++) {
            (
                ,
                string memory opType,
                ,
                ,
                string memory operator,
                uint256 timestamp,
                
            ) = traceEvidence.getHistoryRecord(_traceID, i);
            
            operationTypes[i] = opType;
            operators[i] = operator;
            timestamps[i] = timestamp;
        }
    }
    
    /**
     * @dev 执行审计
     * @param _traceID 追溯码
     * @param _operation 操作类型
     * @param _details 审计详情
     */
    function performAudit(
        string memory _traceID,
        string memory _operation,
        string memory _details
    ) public onlyRegulator {
        AuditLog memory log = AuditLog({
            traceID: _traceID,
            operation: _operation,
            regulator: msg.sender,
            timestamp: block.timestamp,
            details: _details
        });
        
        auditLogs.push(log);
        
        emit AuditPerformed(
            _traceID,
            msg.sender,
            _operation,
            block.timestamp
        );
    }
    
    /**
     * @dev 获取审计日志数量
     * @return 日志数量
     */
    function getAuditLogCount() public view returns (uint256) {
        return auditLogs.length;
    }
    
    /**
     * @dev 获取审计日志
     * @param _index 索引
     * @return 审计日志
     */
    function getAuditLog(uint256 _index)
        public
        view
        returns (
            string memory traceID,
            string memory operation,
            address regulator,
            uint256 timestamp,
            string memory details
        )
    {
        require(_index < auditLogs.length, "Index out of range");
        AuditLog memory log = auditLogs[_index];
        return (
            log.traceID,
            log.operation,
            log.regulator,
            log.timestamp,
            log.details
        );
    }
}
