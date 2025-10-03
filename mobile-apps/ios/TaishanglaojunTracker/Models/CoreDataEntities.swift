//
//  CoreDataEntities.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import Foundation
import CoreData

// MARK: - ChatMessageEntity
@objc(ChatMessageEntity)
public class ChatMessageEntity: NSManagedObject {
    
}

extension ChatMessageEntity {
    
    @nonobjc public class func fetchRequest() -> NSFetchRequest<ChatMessageEntity> {
        return NSFetchRequest<ChatMessageEntity>(entityName: "ChatMessageEntity")
    }
    
    @NSManaged public var id: String?
    @NSManaged public var conversationId: String?
    @NSManaged public var content: String?
    @NSManaged public var messageType: String?
    @NSManaged public var sender: String?
    @NSManaged public var timestamp: Double
    @NSManaged public var status: String?
    @NSManaged public var metadata: String?
    @NSManaged public var conversation: ConversationEntity?
    
}

extension ChatMessageEntity: Identifiable {
    
}

// MARK: - ConversationEntity
@objc(ConversationEntity)
public class ConversationEntity: NSManagedObject {
    
}

extension ConversationEntity {
    
    @nonobjc public class func fetchRequest() -> NSFetchRequest<ConversationEntity> {
        return NSFetchRequest<ConversationEntity>(entityName: "ConversationEntity")
    }
    
    @NSManaged public var id: String?
    @NSManaged public var title: String?
    @NSManaged public var createdAt: Double
    @NSManaged public var updatedAt: Double
    @NSManaged public var lastMessageId: String?
    @NSManaged public var messageCount: Int32
    @NSManaged public var isArchived: Bool
    @NSManaged public var aiPersonality: String?
    @NSManaged public var messages: NSSet?
    
}

// MARK: Generated accessors for messages
extension ConversationEntity {
    
    @objc(addMessagesObject:)
    @NSManaged public func addToMessages(_ value: ChatMessageEntity)
    
    @objc(removeMessagesObject:)
    @NSManaged public func removeFromMessages(_ value: ChatMessageEntity)
    
    @objc(addMessages:)
    @NSManaged public func addToMessages(_ values: NSSet)
    
    @objc(removeMessages:)
    @NSManaged public func removeFromMessages(_ values: NSSet)
    
}

extension ConversationEntity: Identifiable {
    
}

// MARK: - TrajectoryEntity
@objc(TrajectoryEntity)
public class TrajectoryEntity: NSManagedObject {
    
}

extension TrajectoryEntity {
    
    @nonobjc public class func fetchRequest() -> NSFetchRequest<TrajectoryEntity> {
        return NSFetchRequest<TrajectoryEntity>(entityName: "TrajectoryEntity")
    }
    
    @NSManaged public var id: String?
    @NSManaged public var name: String?
    @NSManaged public var startTime: Double
    @NSManaged public var endTime: Double
    @NSManaged public var totalDistance: Double
    @NSManaged public var totalDuration: Double
    @NSManaged public var isRecording: Bool
    @NSManaged public var locationPoints: NSSet?
    
}

// MARK: Generated accessors for locationPoints
extension TrajectoryEntity {
    
    @objc(addLocationPointsObject:)
    @NSManaged public func addToLocationPoints(_ value: LocationPointEntity)
    
    @objc(removeLocationPointsObject:)
    @NSManaged public func removeFromLocationPoints(_ value: LocationPointEntity)
    
    @objc(addLocationPoints:)
    @NSManaged public func addToLocationPoints(_ values: NSSet)
    
    @objc(removeLocationPoints:)
    @NSManaged public func removeFromLocationPoints(_ values: NSSet)
    
}

extension TrajectoryEntity: Identifiable {
    
}

// MARK: - LocationPointEntity
@objc(LocationPointEntity)
public class LocationPointEntity: NSManagedObject {
    
}

extension LocationPointEntity {
    
    @nonobjc public class func fetchRequest() -> NSFetchRequest<LocationPointEntity> {
        return NSFetchRequest<LocationPointEntity>(entityName: "LocationPointEntity")
    }
    
    @NSManaged public var id: String?
    @NSManaged public var latitude: Double
    @NSManaged public var longitude: Double
    @NSManaged public var timestamp: Double
    @NSManaged public var accuracy: Double
    @NSManaged public var altitude: Double
    @NSManaged public var speed: Double
    @NSManaged public var bearing: Double
    @NSManaged public var trajectoryId: String?
    @NSManaged public var trajectory: TrajectoryEntity?
    
}

extension LocationPointEntity: Identifiable {
    
}