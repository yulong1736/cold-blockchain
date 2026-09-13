// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title TraceEvidence
 * @dev 跨境冷链物流溯源存证合约
 */
contract TraceEvidence {
    
    // 存证记录结构
    struct Evidence {
        string traceID;           // 追溯码
        string dataHash;          // 数据哈希
        string operator;          // 操作者
        uint256 timestamp;        // 时间戳
        uint256 blockNumber;      // 区块号
        bool isValid;             // 是否有效
    }
    
    // 操作历史记录结构
    struct HistoryRecord {
        string traceID;           // 追溯码
        string operationType;    // 操作类型: create, update, delete
        string oldHash;           // 旧哈希
        string newHash;           // 新哈希
        string operator;          // 操作者
        uint256 timestamp;        // 时间戳
        string changeDetails;     // 变更详情
    }
    
    // 存储映射
    mapping(string => Evidence) public evidences;           // 追溯码 -> 存证记录
    mapping(string => HistoryRecord[]) public histories;    // 追溯码 -> 历史记录数组
    mapping(string => bool) public traceIDExists;          // 追溯码是否存在
    
    // 事件定义
    event EvidenceStored(
        string indexed traceID,
        string dataHash,
        string operator,
        uint256 timestamp,
        uint256 blockNumber
    );
    
    event HistoryRecorded(
        string indexed traceID,
        string operationType,
        string operator,
        uint256 timestamp
    );

    /// 温控 / 物流等附属存证（按追溯码追加，不覆盖主商品存证 evidences[traceID]）
    struct AuxEvidence {
        string kind;           // 约定: temperature | logistics
        string dataHash;
        string operator;
        uint256 timestamp;
        uint256 blockNumber;
    }

    mapping(string => AuxEvidence[]) private auxEvidencesByTrace;

    event AuxEvidenceStored(
        string indexed traceID,
        string kind,
        string dataHash,
        string operator,
        uint256 timestamp,
        uint256 blockNumber
    );
    
    /**
     * @dev 存储存证信息
     * @param _traceID 追溯码
     * @param _dataHash 数据哈希
     * @param _operator 操作者
     */
    function storeEvidence(
        string memory _traceID,
        string memory _dataHash,
        string memory _operator
    ) public {
        require(bytes(_traceID).length > 0, "TraceID cannot be empty");
        require(bytes(_dataHash).length > 0, "DataHash cannot be empty");
        
        Evidence memory evidence = Evidence({
            traceID: _traceID,
            dataHash: _dataHash,
            operator: _operator,
            timestamp: block.timestamp,
            blockNumber: block.number,
            isValid: true
        });
        
        evidences[_traceID] = evidence;
        traceIDExists[_traceID] = true;
        
        emit EvidenceStored(
            _traceID,
            _dataHash,
            _operator,
            block.timestamp,
            block.number
        );
    }
    
    /**
     * @dev 记录操作历史
     * @param _traceID 追溯码
     * @param _operationType 操作类型
     * @param _oldHash 旧哈希
     * @param _newHash 新哈希
     * @param _operator 操作者
     * @param _changeDetails 变更详情
     */
    function recordHistory(
        string memory _traceID,
        string memory _operationType,
        string memory _oldHash,
        string memory _newHash,
        string memory _operator,
        string memory _changeDetails
    ) public {
        require(bytes(_traceID).length > 0, "TraceID cannot be empty");
        
        HistoryRecord memory record = HistoryRecord({
            traceID: _traceID,
            operationType: _operationType,
            oldHash: _oldHash,
            newHash: _newHash,
            operator: _operator,
            timestamp: block.timestamp,
            changeDetails: _changeDetails
        });
        
        histories[_traceID].push(record);
        
        emit HistoryRecorded(
            _traceID,
            _operationType,
            _operator,
            block.timestamp
        );
    }
    
    /**
     * @dev 查询存证信息
     * @param _traceID 追溯码
     * @return traceID 追溯码
     * @return dataHash 数据哈希
     * @return operator 操作者
     * @return timestamp 时间戳
     * @return blockNumber 区块号
     * @return isValid 是否有效
     */
    function queryEvidence(string memory _traceID) 
        public 
        view 
        returns (
            string memory traceID,
            string memory dataHash,
            string memory operator,
            uint256 timestamp,
            uint256 blockNumber,
            bool isValid
        ) 
    {
        require(traceIDExists[_traceID], "TraceID does not exist");
        
        Evidence memory evidence = evidences[_traceID];
        return (
            evidence.traceID,
            evidence.dataHash,
            evidence.operator,
            evidence.timestamp,
            evidence.blockNumber,
            evidence.isValid
        );
    }
    
    /**
     * @dev 查询操作历史
     * @param _traceID 追溯码
     * @return count 历史记录数组长度
     */
    function getHistoryCount(string memory _traceID) 
        public 
        view 
        returns (uint256 count) 
    {
        return histories[_traceID].length;
    }
    
    /**
     * @dev 获取指定索引的历史记录
     * @param _traceID 追溯码
     * @param _index 索引
     * @return traceID 追溯码
     * @return operationType 操作类型
     * @return oldHash 旧哈希
     * @return newHash 新哈希
     * @return operator 操作者
     * @return timestamp 时间戳
     * @return changeDetails 变更详情
     */
    function getHistoryRecord(string memory _traceID, uint256 _index)
        public
        view
        returns (
            string memory traceID,
            string memory operationType,
            string memory oldHash,
            string memory newHash,
            string memory operator,
            uint256 timestamp,
            string memory changeDetails
        )
    {
        require(_index < histories[_traceID].length, "Index out of range");
        
        HistoryRecord memory record = histories[_traceID][_index];
        return (
            record.traceID,
            record.operationType,
            record.oldHash,
            record.newHash,
            record.operator,
            record.timestamp,
            record.changeDetails
        );
    }
    
    /**
     * @dev 验证追溯码是否存在
     * @param _traceID 追溯码
     * @return exists 是否存在
     */
    function verifyTraceID(string memory _traceID) 
        public 
        view 
        returns (bool exists) 
    {
        exists = traceIDExists[_traceID];
    }
    
    /**
     * @dev 禁用存证（逻辑删除）
     * @param _traceID 追溯码
     */
    function disableEvidence(string memory _traceID) public {
        require(traceIDExists[_traceID], "TraceID does not exist");
        evidences[_traceID].isValid = false;
    }

    /**
     * @dev 追加温控或物流存证（多条记录共用同一 traceID）
     * @param _traceID 追溯码
     * @param _kind temperature | logistics
     * @param _dataHash 业务记录内容哈希
     * @param _operator 操作者标识
     */
    function appendAuxEvidence(
        string memory _traceID,
        string memory _kind,
        string memory _dataHash,
        string memory _operator
    ) public {
        require(bytes(_traceID).length > 0, "TraceID cannot be empty");
        require(bytes(_kind).length > 0, "Kind cannot be empty");
        require(bytes(_dataHash).length > 0, "DataHash cannot be empty");

        AuxEvidence memory row = AuxEvidence({
            kind: _kind,
            dataHash: _dataHash,
            operator: _operator,
            timestamp: block.timestamp,
            blockNumber: block.number
        });

        auxEvidencesByTrace[_traceID].push(row);

        emit AuxEvidenceStored(
            _traceID,
            _kind,
            _dataHash,
            _operator,
            block.timestamp,
            block.number
        );
    }

    function getAuxEvidenceCount(string memory _traceID) public view returns (uint256) {
        return auxEvidencesByTrace[_traceID].length;
    }

    function getAuxEvidenceRecord(string memory _traceID, uint256 _index)
        public
        view
        returns (
            string memory kind,
            string memory dataHash,
            string memory operator,
            uint256 timestamp,
            uint256 blockNumber
        )
    {
        require(_index < auxEvidencesByTrace[_traceID].length, "Index out of range");
        AuxEvidence memory row = auxEvidencesByTrace[_traceID][_index];
        return (row.kind, row.dataHash, row.operator, row.timestamp, row.blockNumber);
    }
}
